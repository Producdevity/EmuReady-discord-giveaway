package application

import (
	"strings"
	"testing"
	"time"
)

func TestOAuthStateRoundTripUsesRandomNonce(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	service := NewOAuthStateService(strings.Repeat("s", 32), time.Minute)
	service.now = func() time.Time { return now }

	first, err := service.CreateState(123)
	if err != nil {
		t.Fatalf("create first state: %v", err)
	}
	second, err := service.CreateState(123)
	if err != nil {
		t.Fatalf("create second state: %v", err)
	}
	if first == second {
		t.Fatal("expected state nonce to make repeated states unique")
	}

	discordID, err := service.VerifyState(first)
	if err != nil {
		t.Fatalf("verify state: %v", err)
	}
	if discordID != 123 {
		t.Fatalf("discord id mismatch: got %d", discordID)
	}
}

func TestOAuthStateRejectsTamperingExpiryAndFutureIssueTime(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	service := NewOAuthStateService(strings.Repeat("s", 32), time.Minute)
	service.now = func() time.Time { return now }

	state, err := service.CreateState(123)
	if err != nil {
		t.Fatalf("create state: %v", err)
	}
	tampered := strings.Replace(state, ".123.", ".456.", 1)
	if _, err := service.VerifyState(tampered); err == nil {
		t.Fatal("expected tampered state to fail")
	}

	service.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := service.VerifyState(state); err == nil {
		t.Fatal("expected expired state to fail")
	}

	service.now = func() time.Time { return now.Add(time.Minute) }
	futureState, err := service.CreateState(123)
	if err != nil {
		t.Fatalf("create future state: %v", err)
	}
	service.now = func() time.Time { return now }
	if _, err := service.VerifyState(futureState); err == nil {
		t.Fatal("expected future-issued state to fail")
	}
}
