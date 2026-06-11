package discord

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/rs/zerolog"
)

func TestVerifySignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	timestamp := "1700000000"
	body := []byte(`{"type":1}`)
	signature := ed25519.Sign(privateKey, append([]byte(timestamp), body...))

	client := NewClient(nil, "token", zerolog.Nop())
	if err := client.VerifySignature(timestamp, hex.EncodeToString(signature), body, hex.EncodeToString(publicKey)); err != nil {
		t.Fatalf("verify valid signature: %v", err)
	}
	if err := client.VerifySignature(timestamp, hex.EncodeToString(signature), []byte(`{"type":2}`), hex.EncodeToString(publicKey)); err == nil {
		t.Fatal("expected invalid body to fail signature verification")
	}
	if err := client.VerifySignature(timestamp, "abc", body, hex.EncodeToString(publicKey)); err == nil {
		t.Fatal("expected malformed signature to fail")
	}
}
