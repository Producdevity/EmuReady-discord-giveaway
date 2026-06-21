package application

import (
	"context"
	"fmt"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage"

	"github.com/rs/zerolog"
)

type resetDiscordClient interface {
	EditOriginalInteractionResponse(ctx context.Context, applicationID string, interactionToken string, body interface{}) error
}

type ResetService struct {
	cfg     *config.Config
	repo    storage.EntrantRepository
	discord resetDiscordClient
	logger  zerolog.Logger
}

func NewResetService(
	cfg *config.Config,
	repo storage.EntrantRepository,
	discordClient resetDiscordClient,
	logger zerolog.Logger,
) *ResetService {
	return &ResetService{cfg: cfg, repo: repo, discord: discordClient, logger: logger}
}

func (s *ResetService) Run(ctx context.Context, interaction domain.Interaction) error {
	entrants, err := s.repo.AllEntrants(ctx)
	if err != nil {
		return err
	}
	if len(entrants) == 0 {
		return s.respond(ctx, interaction.Token, "No giveaway entries to reset.")
	}

	entriesArchived := 0
	archiveFailures := 0
	for _, entrant := range entrants {
		if err := s.repo.SoftDeleteEntrant(ctx, entrant.DiscordID, entrant.GithubID); err != nil {
			archiveFailures++
			s.logger.Warn().Err(err).Int64("discord_id", entrant.DiscordID).Msg("reset entrant archive failed")
			continue
		}
		entriesArchived++
	}

	entryLabel := "entries"
	if entriesArchived == 1 {
		entryLabel = "entry"
	}
	content := fmt.Sprintf("Giveaway reset complete. Archived %d stored %s. Giveaway ping roles were left unchanged.", entriesArchived, entryLabel)
	if archiveFailures > 0 {
		content += fmt.Sprintf(" Entry archive failed for %d user(s).", archiveFailures)
	}
	return s.respond(ctx, interaction.Token, content)
}

func (s *ResetService) respond(ctx context.Context, interactionToken, content string) error {
	payload := domain.WebhookMessageEdit{
		Content: content,
	}
	return s.discord.EditOriginalInteractionResponse(ctx, s.cfg.DiscordApplicationID, interactionToken, payload)
}
