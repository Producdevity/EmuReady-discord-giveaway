package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

const requestIDHeader = "X-Request-ID"

var requestCounter uint64

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(requestIDHeader)
		if id == "" {
			id = generateRequestID()
			c.Set(requestIDHeader, id)
		}
		c.Locals("request_id", id)
		return c.Next()
	}
}

func generateRequestID() string {
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	return fmt.Sprintf("%s-%d", hex.EncodeToString(raw), atomic.AddUint64(&requestCounter, 1)+uint64(time.Now().UnixNano()))
}
