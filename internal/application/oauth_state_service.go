package application

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
)

type OAuthStateService struct {
	signingSecret []byte
	ttl           time.Duration
	now           func() time.Time
}

func NewOAuthStateService(secret string, ttl time.Duration) *OAuthStateService {
	return &OAuthStateService{signingSecret: []byte(secret), ttl: ttl, now: time.Now}
}

func (s *OAuthStateService) CreateState(discordID int64) (string, error) {
	if discordID <= 0 {
		return "", domain.ErrStateInvalid
	}
	nonceRaw := make([]byte, 32)
	if _, err := rand.Read(nonceRaw); err != nil {
		return "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceRaw)
	issuedAt := s.now().Unix()
	payload := fmt.Sprintf("v1.%d.%d.%s", discordID, issuedAt, nonce)
	h := hmac.New(sha256.New, s.signingSecret)
	_, _ = h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s.%s", payload, signature), nil
}

func (s *OAuthStateService) VerifyState(raw string) (int64, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 5 || parts[0] != "v1" {
		return 0, domain.ErrStateInvalid
	}
	discordID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, domain.ErrStateInvalid
	}
	if discordID <= 0 {
		return 0, domain.ErrStateInvalid
	}
	issuedAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, domain.ErrStateInvalid
	}
	if issuedAt <= 0 {
		return 0, domain.ErrStateInvalid
	}
	issuedAtTime := time.Unix(issuedAt, 0)
	now := s.now()
	if issuedAtTime.After(now.Add(30 * time.Second)) {
		return 0, domain.ErrStateInvalid
	}
	if now.Sub(issuedAtTime) > s.ttl {
		return 0, domain.ErrStateInvalid
	}
	if len(parts[3]) < 32 {
		return 0, domain.ErrStateInvalid
	}
	if _, err := base64.RawURLEncoding.DecodeString(parts[3]); err != nil {
		return 0, domain.ErrStateInvalid
	}

	payload := fmt.Sprintf("v1.%d.%d.%s", discordID, issuedAt, parts[3])
	h := hmac.New(sha256.New, s.signingSecret)
	_, _ = h.Write([]byte(payload))
	expected := hex.EncodeToString(h.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(parts[4]))) {
		return 0, domain.ErrStateInvalid
	}
	return discordID, nil
}
