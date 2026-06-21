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
const maxRequestIDLength = 128

var requestCounter uint64

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(requestIDHeader)
		if !validRequestID(id) {
			id = generateRequestID()
		}
		c.Set(requestIDHeader, id)
		c.Locals("request_id", id)
		return c.Next()
	}
}

func generateRequestID() string {
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	return fmt.Sprintf("%s-%d", hex.EncodeToString(raw), atomic.AddUint64(&requestCounter, 1)+uint64(time.Now().UnixNano()))
}

func validRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.' || r == ':':
		default:
			return false
		}
	}
	return true
}
