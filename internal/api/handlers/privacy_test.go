package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestPrivacyHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/privacy", NewPrivacyHandler())

	req := httptest.NewRequest(http.MethodGet, "/privacy", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("privacy request failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if got := res.Header.Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content type = %q, want text/html", got)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	html := string(body)
	for _, want := range []string{
		"Privacy Policy | EmuReady Discord Giveaway",
		"Discord user ID",
		"GitHub user ID and username",
		"Your Rights and Choices",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("privacy page missing %q", want)
		}
	}
}
