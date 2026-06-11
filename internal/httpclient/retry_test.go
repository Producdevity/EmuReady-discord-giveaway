package httpclient

import (
	"net/url"
	"testing"
)

func TestSanitizedURLRedactsWebhookTokenAndQuery(t *testing.T) {
	raw, err := url.Parse("https://discord.com/api/v10/webhooks/app-id/interaction-token/messages/@original?wait=true")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	got := sanitizedURL(raw)
	want := "https://discord.com/api/v10/webhooks/app-id/redacted/messages/@original"
	if got != want {
		t.Fatalf("sanitized url mismatch:\nwant %s\n got %s", want, got)
	}
}
