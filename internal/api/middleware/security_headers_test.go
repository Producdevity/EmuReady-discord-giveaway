package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.Type("html").SendString("<!doctype html><title>ok</title>")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	assertHeader(t, res, "Cache-Control", "no-store")
	assertHeader(t, res, "Referrer-Policy", "no-referrer")
	assertHeader(t, res, "X-Content-Type-Options", "nosniff")
	assertHeader(t, res, "X-Frame-Options", "DENY")
	assertHeader(t, res, "Cross-Origin-Resource-Policy", "same-origin")
	assertHeader(t, res, "Permissions-Policy", "camera=(), geolocation=(), microphone=()")

	csp := res.Header.Get("Content-Security-Policy")
	for _, want := range []string{
		"default-src 'none'",
		"script-src 'none'",
		"frame-ancestors 'none'",
		"style-src 'unsafe-inline'",
	} {
		if !strings.Contains(csp, want) {
			t.Fatalf("content security policy = %q, missing %q", csp, want)
		}
	}
}

func assertHeader(t *testing.T, res *http.Response, key, want string) {
	t.Helper()
	if got := res.Header.Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
