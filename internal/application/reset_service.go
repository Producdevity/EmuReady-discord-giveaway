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
	RemoveRoleFromMember(ctx context.Context, guildID, userID, roleID string) error
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

	rolesRemoved := 0
	roleFailures := 0
	entriesDeleted := 0
	deleteFailures := 0
	for _, entrant := range entrants {
		discordID := fmt.Sprintf("%d", entrant.DiscordID)
		if err := s.discord.RemoveRoleFromMember(ctx, s.cfg.GuildID, discordID, s.cfg.GiveawayRoleID); err != nil {
			roleFailures++
			s.logger.Warn().Err(err).Int64("discord_id", entrant.DiscordID).Msg("reset role removal failed")
		} else {
			rolesRemoved++
		}

		if err := s.repo.DeleteEntrant(ctx, entrant.DiscordID, entrant.GithubID); err != nil {
			deleteFailures++
			s.logger.Warn().Err(err).Int64("discord_id", entrant.DiscordID).Msg("reset entrant delete failed")
			continue
		}
		entriesDeleted++
	}

	entryLabel := "entries"
	if entriesDeleted == 1 {
		entryLabel = "entry"
	}
	content := fmt.Sprintf("Giveaway reset complete. Deleted %d stored %s. Removed the giveaway role from %d user(s).", entriesDeleted, entryLabel, rolesRemoved)
	if roleFailures > 0 || deleteFailures > 0 {
		content += fmt.Sprintf(" Role removal failed for %d user(s); entry deletion failed for %d user(s).", roleFailures, deleteFailures)
	}
	return s.respond(ctx, interaction.Token, content)
}

func (s *ResetService) respond(ctx context.Context, interactionToken, content string) error {
	payload := domain.WebhookMessageEdit{
		Content: content,
	}
	return s.discord.EditOriginalInteractionResponse(ctx, s.cfg.DiscordApplicationID, interactionToken, payload)
}
