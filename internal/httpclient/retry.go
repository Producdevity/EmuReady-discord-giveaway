package httpclient

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var defaultRetryPolicy = RetryPolicy{
	MaxAttempts: 3,
	BaseDelay:   200 * time.Millisecond,
	MaxDelay:    1200 * time.Millisecond,
	Jitter:      75 * time.Millisecond,
}

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      time.Duration
}

// DefaultRetryPolicy returns a production-safe baseline for HTTP calls.
func DefaultRetryPolicy() RetryPolicy {
	return defaultRetryPolicy
}

func DoWithRetry(
	ctx context.Context,
	client *http.Client,
	requestFactory func() (*http.Request, error),
	logger zerolog.Logger,
	policy RetryPolicy,
) (*http.Response, error) {
	policy = normalizeRetryPolicy(policy)
	var lastErr error
	var response *http.Response

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		req, err := requestFactory()
		if err != nil {
			return nil, err
		}

		response, lastErr = client.Do(req.WithContext(ctx))
		if lastErr == nil && !isRetryableStatus(response.StatusCode) {
			return response, nil
		}

		if !shouldRetry(response, lastErr, policy.MaxAttempts, attempt) {
			return response, lastErr
		}

		if response != nil {
			if body, readErr := io.ReadAll(io.LimitReader(response.Body, 2048)); readErr == nil {
				_ = response.Body.Close()
				logger.Warn().
					Err(lastErr).
					Int("status", response.StatusCode).
					Dur("status_retry_after", getRetryAfter(response.Header)).
					Str("status_body", strings.TrimSpace(string(body))).
					Int("attempt", attempt).
					Str("url", sanitizedURL(req.URL)).
					Msg("upstream request retryable; retrying")
			}
		} else {
			logger.Warn().Err(lastErr).Int("attempt", attempt).Str("url", sanitizedURL(req.URL)).Msg("upstream request retryable; retrying")
		}

		delay := nextDelay(policy, attempt)
		responseStatus := time.Duration(0)
		if response != nil {
			responseStatus = getRetryAfter(response.Header)
		}
		if responseStatus > 0 {
			delay = maxDuration(delay, responseStatus)
		}
		if !sleepWithContext(ctx, delay) {
			return nil, ctx.Err()
		}
	}

	return response, lastErr
}

func normalizeRetryPolicy(policy RetryPolicy) RetryPolicy {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = 250 * time.Millisecond
	}
	if policy.MaxDelay < policy.BaseDelay {
		policy.MaxDelay = policy.BaseDelay
	}
	if policy.Jitter < 0 {
		policy.Jitter = 0
	}
	return policy
}

func shouldRetry(response *http.Response, err error, maxAttempts, attempt int) bool {
	if attempt >= maxAttempts {
		return false
	}
	if err != nil {
		return true
	}
	if response == nil {
		return false
	}
	return isRetryableStatus(response.StatusCode)
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || (status >= 500 && status < 600)
}

func nextDelay(policy RetryPolicy, attempt int) time.Duration {
	attempt = maxInt(attempt-1, 0)
	delay := policy.BaseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
			break
		}
	}
	if policy.Jitter <= 0 {
		return delay
	}
	return delay + time.Duration(rand.Int63n(int64(policy.Jitter)+1))
}

func getRetryAfter(header http.Header) time.Duration {
	raw := header.Get("Retry-After")
	if raw == "" {
		return 0
	}

	raw = strings.TrimSpace(raw)
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	expiresAt, err := http.ParseTime(raw)
	if err != nil {
		return 0
	}
	delay := time.Until(expiresAt)
	if delay < 0 {
		return 0
	}
	return delay
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sanitizedURL(raw *url.URL) string {
	if raw == nil {
		return ""
	}
	copied := *raw
	copied.RawQuery = ""
	parts := strings.Split(copied.Path, "/")
	for i, part := range parts {
		if part == "webhooks" && i+2 < len(parts) {
			parts[i+2] = "redacted"
			break
		}
	}
	copied.Path = strings.Join(parts, "/")
	return copied.String()
}
