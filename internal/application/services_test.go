package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"

	"github.com/rs/zerolog"
)

func TestCallbackRollsBackEntrantWhenRoleAssignmentFails(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	state := NewOAuthStateService(strings.Repeat("s", 32), time.Minute)
	state.now = func() time.Time { return now }
	rawState, err := state.CreateState(42)
	if err != nil {
		t.Fatalf("create state: %v", err)
	}

	roleErr := errors.New("discord role failed")
	store := &fakeEntrantStore{}
	service, err := NewCallbackService(
		testConfig(),
		state,
		store,
		&fakeCallbackDiscord{addErr: roleErr},
		&fakeCallbackGitHub{user: domain.GitHubUser{ID: 99, Login: "octo"}, hasStar: true},
		zerolog.Nop(),
	)
	if err != nil {
		t.Fatalf("new callback service: %v", err)
	}

	if _, err := service.Handle(context.Background(), "code", rawState); !errors.Is(err, roleErr) {
		t.Fatalf("expected role error, got %v", err)
	}
	if len(store.upserts) != 1 {
		t.Fatalf("expected entrant upsert before role assignment, got %d", len(store.upserts))
	}
	if len(store.deletes) != 1 {
		t.Fatalf("expected rollback delete, got %d", len(store.deletes))
	}
	if store.deletes[0].DiscordID != 42 || store.deletes[0].GithubID != 99 {
		t.Fatalf("rollback target mismatch: %+v", store.deletes[0])
	}
}

func TestWinnerEditsOriginalInteractionWithWebhookPayload(t *testing.T) {
	store := &fakeEntrantStore{
		entrants: []domain.Entrant{
			{DiscordID: 1, GithubID: 11, GithubLogin: "alice"},
			{DiscordID: 2, GithubID: 22, GithubLogin: "bob"},
		},
	}
	discord := &fakeWinnerDiscord{}
	github := &fakeWinnerGitHub{starred: map[string]bool{"alice": true, "bob": false}}

	service, err := NewWinnerService(testConfig(), store, discord, github, zerolog.Nop())
	if err != nil {
		t.Fatalf("new winner service: %v", err)
	}
	if err := service.Run(context.Background(), domain.Interaction{Token: "interaction-token"}, 2); err != nil {
		t.Fatalf("run winner service: %v", err)
	}

	if len(discord.removedUsers) != 1 || discord.removedUsers[0] != "2" {
		t.Fatalf("expected bob role removal, got %v", discord.removedUsers)
	}
	if len(store.deletes) != 1 || store.deletes[0].DiscordID != 2 || store.deletes[0].GithubID != 22 {
		t.Fatalf("expected bob entry deletion, got %v", store.deletes)
	}
	if len(discord.edits) != 1 {
		t.Fatalf("expected one edit, got %d", len(discord.edits))
	}
	payload, ok := discord.edits[0].(domain.WebhookMessageEdit)
	if !ok {
		t.Fatalf("expected webhook message edit payload, got %T", discord.edits[0])
	}
	if !strings.Contains(payload.Content, "<@1> (alice)") {
		t.Fatalf("winner content missing alice mention: %q", payload.Content)
	}
	if payload.AllowedMentions == nil || len(payload.AllowedMentions.Parse) != 1 || payload.AllowedMentions.Parse[0] != "users" {
		t.Fatalf("allowed mentions mismatch: %+v", payload.AllowedMentions)
	}
}

func testConfig() *config.Config {
	return &config.Config{
		DiscordApplicationID: "app-id",
		GuildID:              "guild-id",
		GiveawayRoleID:       "role-id",
		GithubRepo:           "owner/repo",
		BaseURL:              "https://example.com",
		WinnerDefaultCount:   1,
		WinnerMax:            10,
		WinnerConcurrency:    2,
	}
}

type fakeEntrantStore struct {
	entrants []domain.Entrant
	upserts  []domain.Entrant
	deletes  []domain.Entrant
}

func (s *fakeEntrantStore) FindEntrant(context.Context, int64) (*domain.Entrant, error) {
	return nil, domain.ErrNotFound
}

func (s *fakeEntrantStore) UpsertEntrant(_ context.Context, entrant domain.Entrant) error {
	s.upserts = append(s.upserts, entrant)
	return nil
}

func (s *fakeEntrantStore) DeleteEntrant(_ context.Context, discordID int64, githubID int64) error {
	s.deletes = append(s.deletes, domain.Entrant{DiscordID: discordID, GithubID: githubID})
	return nil
}

func (s *fakeEntrantStore) CountEntrants(context.Context) (int, error) {
	return len(s.entrants), nil
}

func (s *fakeEntrantStore) AllEntrants(context.Context) ([]domain.Entrant, error) {
	return s.entrants, nil
}

type fakeCallbackDiscord struct {
	addErr error
}

func (d *fakeCallbackDiscord) AddRoleToMember(context.Context, string, string, string) error {
	return d.addErr
}

type fakeCallbackGitHub struct {
	user    domain.GitHubUser
	hasStar bool
}

func (g *fakeCallbackGitHub) ExchangeCode(context.Context, string, string) (string, error) {
	return "token", nil
}

func (g *fakeCallbackGitHub) GetCurrentUser(context.Context, string) (domain.GitHubUser, error) {
	return g.user, nil
}

func (g *fakeCallbackGitHub) HasStarredRepo(context.Context, string, string, string) (bool, error) {
	return g.hasStar, nil
}

type fakeWinnerDiscord struct {
	removedUsers []string
	edits        []interface{}
}

func (d *fakeWinnerDiscord) RemoveRoleFromMember(_ context.Context, _ string, userID string, _ string) error {
	d.removedUsers = append(d.removedUsers, userID)
	return nil
}

func (d *fakeWinnerDiscord) EditOriginalInteractionResponse(_ context.Context, _ string, _ string, body interface{}) error {
	d.edits = append(d.edits, body)
	return nil
}

type fakeWinnerGitHub struct {
	starred map[string]bool
}

func (g *fakeWinnerGitHub) CheckUsersStar(_ context.Context, usernames []string, _ string, _ string, _ int) (map[string]bool, bool, error) {
	result := make(map[string]bool, len(usernames))
	for _, username := range usernames {
		result[strings.ToLower(username)] = g.starred[strings.ToLower(username)]
	}
	return result, true, nil
}
