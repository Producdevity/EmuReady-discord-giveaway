package middleware

import "github.com/gofiber/fiber/v2"

const contentSecurityPolicy = "default-src 'none'; base-uri 'none'; connect-src 'none'; form-action 'none'; frame-ancestors 'none'; img-src 'none'; object-src 'none'; script-src 'none'; style-src 'unsafe-inline'"

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store")
		c.Set("Content-Security-Policy", contentSecurityPolicy)
		c.Set("Cross-Origin-Resource-Policy", "same-origin")
		c.Set("Permissions-Policy", "camera=(), geolocation=(), microphone=()")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		return c.Next()
	}
}
