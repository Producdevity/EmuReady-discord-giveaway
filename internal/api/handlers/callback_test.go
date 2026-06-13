package handlers

import (
	"errors"
	"strings"
	"testing"
)

func TestLookupCallbackFailureForMissingStar(t *testing.T) {
	failure := lookupCallbackFailure(errors.New("user does not star Producdevity/EmuReady"), "Producdevity/EmuReady")

	if !failure.ShowRepoLink {
		t.Fatal("expected missing-star failure to show repository link")
	}
	if !strings.Contains(failure.Message, "Producdevity/EmuReady") {
		t.Fatalf("expected repo name in message, got %q", failure.Message)
	}
	if !strings.Contains(failure.NextStep, "/enter-giveaway") {
		t.Fatalf("expected retry command in next step, got %q", failure.NextStep)
	}
}

func TestLookupCallbackFailureForMissingDiscordPermissions(t *testing.T) {
	failure := lookupCallbackFailure(errors.New(`discord api error status=403 body={"message":"Missing Permissions","code":50013}`), "Producdevity/EmuReady")

	if failure.ShowRepoLink {
		t.Fatal("did not expect Discord permission failure to show repository link")
	}
	if strings.Contains(strings.ToLower(failure.NextStep), "star") {
		t.Fatalf("did not expect star instruction in next step, got %q", failure.NextStep)
	}
	if !strings.Contains(failure.Message, "Discord") {
		t.Fatalf("expected Discord-specific message, got %q", failure.Message)
	}
}
