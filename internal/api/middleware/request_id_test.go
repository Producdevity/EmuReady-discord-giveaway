package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDAcceptsValidHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("request_id").(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(requestIDHeader, "trace-123_ABC.456:789")
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if got := res.Header.Get(requestIDHeader); got != "trace-123_ABC.456:789" {
		t.Fatalf("response request id = %q", got)
	}
}

func TestRequestIDReplacesInvalidHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("request_id").(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(requestIDHeader, strings.Repeat("x", maxRequestIDLength+1))
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	got := res.Header.Get(requestIDHeader)
	if got == "" {
		t.Fatal("response request id is empty")
	}
	if len(got) > maxRequestIDLength {
		t.Fatalf("response request id length = %d, want <= %d", len(got), maxRequestIDLength)
	}
	if got == strings.Repeat("x", maxRequestIDLength+1) {
		t.Fatal("invalid request id was reused")
	}
}
