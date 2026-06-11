package discord

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/producdevity/emuready-discord-giveaway/internal/httpclient"

	"github.com/rs/zerolog"
)

const discordAPIPrefix = "https://discord.com/api/v10"

type Client struct {
	httpClient  *http.Client
	botToken    string
	logger      zerolog.Logger
	retryPolicy httpclient.RetryPolicy
}

func NewClient(httpClient *http.Client, botToken string, logger zerolog.Logger) *Client {
	return &Client{
		httpClient:  httpClient,
		botToken:    botToken,
		logger:      logger,
		retryPolicy: httpclient.DefaultRetryPolicy(),
	}
}

func (c *Client) SetRetryPolicy(policy httpclient.RetryPolicy) {
	c.retryPolicy = policy
}

func (c *Client) buildRequest(ctx context.Context, method, endpoint string, payload any) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		if payloadMap, ok := payload.(map[string]interface{}); ok && len(payloadMap) == 0 {
			body = bytes.NewBufferString("{}")
		} else {
			b, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(b)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "emuready-discord-giveaway")
	return req, nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, payload any) (*http.Response, error) {
	return httpclient.DoWithRetry(ctx, c.httpClient, func() (*http.Request, error) {
		return c.buildRequest(ctx, method, endpoint, payload)
	}, c.logger, c.retryPolicy)
}

type discordMember struct {
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	Roles []string `json:"roles"`
}

type applicationInfo struct {
	ID string `json:"id"`
}

func (c *Client) VerifySignature(timestamp, signature string, body []byte, publicKey string) error {
	if signature == "" || timestamp == "" {
		return fmt.Errorf("missing signature headers")
	}
	key, err := hex.DecodeString(strings.TrimSpace(publicKey))
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if len(key) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key length")
	}
	sig, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature length")
	}
	msg := append([]byte(timestamp), body...)
	if !ed25519.Verify(key, msg, sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func (c *Client) GetMembersByRole(ctx context.Context, guildID string, roleID string) ([]string, error) {
	if strings.TrimSpace(guildID) == "" || strings.TrimSpace(roleID) == "" {
		return nil, fmt.Errorf("missing guildID or roleID")
	}
	memberIDs := make([]string, 0)
	after := ""
	for {
		url := fmt.Sprintf("%s/guilds/%s/members?limit=%d", discordAPIPrefix, guildID, 1000)
		if after != "" {
			url += "&after=" + after
		}
		var members []discordMember
		if err := c.request(ctx, http.MethodGet, url, nil, &members); err != nil {
			return nil, err
		}
		for _, member := range members {
			if member.User.ID == "" {
				continue
			}
			for _, role := range member.Roles {
				if role == roleID {
					memberIDs = append(memberIDs, member.User.ID)
					break
				}
			}
		}
		if len(members) < 1000 {
			break
		}
		if len(members) > 0 {
			after = members[len(members)-1].User.ID
		}
	}
	return memberIDs, nil
}

func (c *Client) AddRoleToMember(ctx context.Context, guildID, userID, roleID string) error {
	url := fmt.Sprintf("%s/guilds/%s/members/%s/roles/%s", discordAPIPrefix, guildID, userID, roleID)
	return c.request(ctx, http.MethodPut, url, map[string]interface{}{}, nil)
}

func (c *Client) RemoveRoleFromMember(ctx context.Context, guildID, userID, roleID string) error {
	url := fmt.Sprintf("%s/guilds/%s/members/%s/roles/%s", discordAPIPrefix, guildID, userID, roleID)
	return c.request(ctx, http.MethodDelete, url, nil, nil)
}

func (c *Client) EditOriginalInteractionResponse(ctx context.Context, applicationID string, interactionToken string, body interface{}) error {
	url := fmt.Sprintf("%s/webhooks/%s/%s/messages/@original", discordAPIPrefix, applicationID, interactionToken)
	return c.request(ctx, http.MethodPatch, url, body, nil)
}

func (c *Client) GetApplicationID(ctx context.Context) (string, error) {
	var app applicationInfo
	if err := c.request(ctx, http.MethodGet, discordAPIPrefix+"/oauth2/applications/@me", nil, &app); err != nil {
		return "", err
	}
	if app.ID == "" {
		return "", fmt.Errorf("empty application id")
	}
	return app.ID, nil
}

func (c *Client) RegisterGuildCommands(ctx context.Context, applicationID string, guildID string, commands interface{}) error {
	url := fmt.Sprintf("%s/applications/%s/guilds/%s/commands", discordAPIPrefix, applicationID, guildID)
	return c.request(ctx, http.MethodPut, url, commands, nil)
}

func (c *Client) request(ctx context.Context, method, endpoint string, payload any, out any) error {
	resp, err := c.doRequest(ctx, method, endpoint, payload)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
		if err != nil {
			return fmt.Errorf("discord api error status=%d", resp.StatusCode)
		}
		return fmt.Errorf("discord api error status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}
