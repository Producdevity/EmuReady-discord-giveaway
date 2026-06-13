package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestTermsHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/terms", NewTermsHandler())

	req := httptest.NewRequest(http.MethodGet, "/terms", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("terms request failed: %v", err)
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
		"Terms of Service | EmuReady Discord Giveaway",
		"Giveaway Entry and Fair Use",
		"Winner Selection",
		"not sponsored, endorsed, or administered",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("terms page missing %q", want)
		}
	}
}
