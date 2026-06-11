package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/httpclient"

	"github.com/rs/zerolog"
)

const githubAPIPrefix = "https://api.github.com"
const githubOAuthTokenEndpoint = "https://github.com/login/oauth/access_token"

type currentUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	apiToken     string
	apiBaseURL   string
	oauthURL     string
	logger       zerolog.Logger
	retryPolicy  httpclient.RetryPolicy
}

func NewClient(httpClient *http.Client, clientID, clientSecret, apiToken string) *Client {
	return &Client{
		httpClient:   httpClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		apiToken:     apiToken,
		apiBaseURL:   githubAPIPrefix,
		oauthURL:     githubOAuthTokenEndpoint,
		logger:       zerolog.Nop(),
		retryPolicy:  httpclient.DefaultRetryPolicy(),
	}
}

func newClientWithEndpoints(httpClient *http.Client, clientID, clientSecret, apiToken, apiBaseURL, oauthURL string) *Client {
	client := NewClient(httpClient, clientID, clientSecret, apiToken)
	client.apiBaseURL = strings.TrimRight(apiBaseURL, "/")
	client.oauthURL = oauthURL
	return client
}

func (c *Client) SetRetryPolicy(policy httpclient.RetryPolicy) {
	c.retryPolicy = policy
}

func (c *Client) buildRequest(ctx context.Context, method, endpoint string, payload any, accessToken string) (*http.Request, error) {
	var body io.Reader
	contentType := ""
	if payload != nil {
		if values, ok := payload.(url.Values); ok {
			body = strings.NewReader(values.Encode())
			contentType = "application/x-www-form-urlencoded"
		} else {
			b, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(b)
			contentType = "application/json"
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "emuready-discord-giveaway")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	return req, nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, payload any, accessToken string) (*http.Response, error) {
	return httpclient.DoWithRetry(ctx, c.httpClient, func() (*http.Request, error) {
		return c.buildRequest(ctx, method, endpoint, payload, accessToken)
	}, c.logger, c.retryPolicy)
}

func (c *Client) ExchangeCode(ctx context.Context, code string, redirectURI string) (string, error) {
	payload := url.Values{
		"client_id":     []string{c.clientID},
		"client_secret": []string{c.clientSecret},
		"code":          []string{code},
		"redirect_uri":  []string{redirectURI},
	}

	resp, err := c.doRequest(ctx, http.MethodPost, c.oauthURL, payload, "")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("github exchange failed: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	if tr.Error != "" {
		if tr.Description != "" {
			return "", fmt.Errorf("github exchange failed: %s: %s", tr.Error, tr.Description)
		}
		return "", fmt.Errorf("github exchange failed: %s", tr.Error)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("github returned empty access token")
	}
	return tr.AccessToken, nil
}

func (c *Client) GetCurrentUser(ctx context.Context, accessToken string) (domain.GitHubUser, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, c.apiBaseURL+"/user", nil, accessToken)
	if err != nil {
		return domain.GitHubUser{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return domain.GitHubUser{}, fmt.Errorf("github user fetch failed: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var user currentUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return domain.GitHubUser{}, err
	}
	return domain.GitHubUser{ID: user.ID, Login: user.Login}, nil
}

func (c *Client) HasStarredRepo(ctx context.Context, accessToken string, owner, repo string) (bool, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("%s/user/starred/%s/%s", c.apiBaseURL, pathEscape(owner), pathEscape(repo)), nil, accessToken)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return false, fmt.Errorf("github starred check unauthorized: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return false, fmt.Errorf("github starred check failed: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func (c *Client) CheckUsersStar(ctx context.Context, usernames []string, owner string, repo string, concurrency int) (map[string]bool, bool, error) {
	result := make(map[string]bool, len(usernames))
	pending := make(map[string]struct{}, len(usernames))
	for _, login := range usernames {
		l := strings.ToLower(strings.TrimSpace(login))
		if l == "" {
			continue
		}
		result[l] = false
		pending[l] = struct{}{}
	}
	if len(pending) == 0 {
		return result, true, nil
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(pending) {
		concurrency = len(pending)
	}

	jobs := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for login := range jobs {
				starred, err := c.HasUserStarredRepo(ctx, login, owner, repo)
				mu.Lock()
				if err != nil && firstErr == nil {
					firstErr = err
				}
				if err == nil {
					result[login] = starred
				}
				mu.Unlock()
			}
		}()
	}

	for login := range pending {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return result, false, ctx.Err()
		case jobs <- login:
		}
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return result, false, firstErr
	}
	return result, true, nil
}

func (c *Client) HasUserStarredRepo(ctx context.Context, username, owner, repo string) (bool, error) {
	endpoint := fmt.Sprintf("%s/users/%s/starred/%s/%s", c.apiBaseURL, pathEscape(username), pathEscape(owner), pathEscape(repo))
	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil, c.apiToken)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return false, fmt.Errorf("github user starred check failed: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func pathEscape(value string) string {
	return url.PathEscape(strings.TrimSpace(value))
}
