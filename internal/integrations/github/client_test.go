package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestExchangeCodeRequestsJSONAndFormEncodesPayload(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/login/oauth/access_token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); !strings.Contains(got, "application/vnd.github+json") {
			t.Fatalf("missing json accept header: %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("content type mismatch: %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("client_id"); got != "client-id" {
			t.Fatalf("client_id mismatch: %q", got)
		}
		if got := r.Form.Get("client_secret"); got != "client-secret" {
			t.Fatalf("client_secret mismatch: %q", got)
		}
		if got := r.Form.Get("code"); got != "code-123" {
			t.Fatalf("code mismatch: %q", got)
		}
		body, err := json.Marshal(tokenResponse{AccessToken: "token-123"})
		if err != nil {
			t.Fatalf("marshal token response: %v", err)
		}
		return jsonResponse(http.StatusOK, string(body)), nil
	})}

	client := newClientWithEndpoints(httpClient, "client-id", "client-secret", "", "https://github.test", "https://github.test/login/oauth/access_token")
	token, err := client.ExchangeCode(context.Background(), "code-123", "https://example.com/callback")
	if err != nil {
		t.Fatalf("exchange code: %v", err)
	}
	if token != "token-123" {
		t.Fatalf("token mismatch: %q", token)
	}
}

func TestCheckUsersStarUsesRepoStargazersEndpoint(t *testing.T) {
	var mu sync.Mutex
	requests := make([]string, 0, 1)
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		requests = append(requests, r.URL.RequestURI())
		mu.Unlock()
		if got := r.Header.Get("Authorization"); got != "Bearer api-token" {
			t.Fatalf("authorization mismatch: %q", got)
		}
		if r.URL.Path != "/repos/owner/repo/stargazers" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("per_page"); got != "100" {
			t.Fatalf("per_page mismatch: %q", got)
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Fatalf("page mismatch: %q", got)
		}
		return jsonResponse(http.StatusOK, `[{"login":"Alice"}]`), nil
	})}

	client := newClientWithEndpoints(httpClient, "client-id", "client-secret", "api-token", "https://github.test", "https://github.test/login/oauth/access_token")
	starred, completed, err := client.CheckUsersStar(context.Background(), []string{"Alice", "bob"}, "owner", "repo", 2)
	if err != nil {
		t.Fatalf("check users star: %v", err)
	}
	if !completed {
		t.Fatal("expected completed star check")
	}
	if !starred["alice"] {
		t.Fatal("expected alice to be starred")
	}
	if starred["bob"] {
		t.Fatal("expected bob not to be starred")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 1 {
		t.Fatalf("expected 1 stargazer request, got %d", len(requests))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
