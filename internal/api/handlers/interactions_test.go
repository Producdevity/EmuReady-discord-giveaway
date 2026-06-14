package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func TestEntrantsCommandCountsStoredEntries(t *testing.T) {
	handler := NewInteractionHandler(
		&config.Config{},
		nil,
		nil,
		&fakeEntrantCounter{count: 7},
		&fakeInteractionDiscord{},
		nil,
		zerolog.Nop(),
	)

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return handler.handleCommand(c, domain.Interaction{
			Data:    &domain.InteractionData{Name: "entrants"},
			GuildID: "guild-id",
			Token:   "interaction-token",
			Member:  &domain.InteractionMember{User: &domain.DiscordUser{ID: "42"}, Permissions: "32"},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("entrants request failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", res.StatusCode, http.StatusOK)
	}

	var response domain.InteractionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data == nil || response.Data.Content != "Current entrants: 7" {
		t.Fatalf("unexpected response data: %+v", response.Data)
	}
	if response.Data.Flags != domain.MessageFlagEphemeral {
		t.Fatalf("flags = %d, want ephemeral flag", response.Data.Flags)
	}
}

func TestEntrantsCommandRequiresManageServer(t *testing.T) {
	handler := NewInteractionHandler(
		&config.Config{},
		nil,
		nil,
		&fakeEntrantCounter{count: 7},
		&fakeInteractionDiscord{},
		nil,
		zerolog.Nop(),
	)

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return handler.handleCommand(c, domain.Interaction{
			Data:    &domain.InteractionData{Name: "entrants"},
			GuildID: "guild-id",
			Token:   "interaction-token",
			Member:  &domain.InteractionMember{User: &domain.DiscordUser{ID: "42"}, Permissions: "0"},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("entrants request failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", res.StatusCode, http.StatusOK)
	}

	var response domain.InteractionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data == nil || response.Data.Content != "You need Manage Server permission to use /entrants." {
		t.Fatalf("unexpected response data: %+v", response.Data)
	}
	if response.Data.Flags != domain.MessageFlagEphemeral {
		t.Fatalf("flags = %d, want ephemeral flag", response.Data.Flags)
	}
}

type fakeEntrantCounter struct {
	count int
}

func (c *fakeEntrantCounter) CountEntrants(context.Context) (int, error) {
	return c.count, nil
}

type fakeInteractionDiscord struct{}

func (d *fakeInteractionDiscord) VerifySignature(string, string, []byte, string) error {
	return nil
}
