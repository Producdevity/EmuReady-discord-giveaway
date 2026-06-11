package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage"

	"github.com/rs/zerolog"
)

type winnerDiscordClient interface {
	GetMembersByRole(ctx context.Context, guildID string, roleID string) ([]string, error)
	RemoveRoleFromMember(ctx context.Context, guildID, userID, roleID string) error
	EditOriginalInteractionResponse(ctx context.Context, applicationID string, interactionToken string, body interface{}) error
}

type winnerGitHubClient interface {
	CheckUsersStar(ctx context.Context, usernames []string, owner string, repo string, concurrency int) (map[string]bool, bool, error)
}

type WinnerService struct {
	cfg      *config.Config
	repo     storage.EntrantRepository
	discord  winnerDiscordClient
	github   winnerGitHubClient
	logger   zerolog.Logger
	owner    string
	repoName string
}

func NewWinnerService(
	cfg *config.Config,
	repo storage.EntrantRepository,
	discordClient winnerDiscordClient,
	githubClient winnerGitHubClient,
	logger zerolog.Logger,
) (*WinnerService, error) {
	owner, repoName, err := cfg.GitHubRepoParts()
	if err != nil {
		return nil, err
	}
	return &WinnerService{cfg: cfg, repo: repo, discord: discordClient, github: githubClient, logger: logger, owner: owner, repoName: repoName}, nil
}

func (s *WinnerService) Run(ctx context.Context, interaction domain.Interaction, winnerCount int) error {
	if winnerCount <= 0 {
		winnerCount = s.cfg.WinnerDefaultCount
	}
	if winnerCount > s.cfg.WinnerMax {
		winnerCount = s.cfg.WinnerMax
	}

	roleMembers, err := s.discord.GetMembersByRole(ctx, s.cfg.GuildID, s.cfg.GiveawayRoleID)
	if err != nil {
		return err
	}
	memberSet := make(map[string]struct{}, len(roleMembers))
	for _, memberID := range roleMembers {
		memberSet[memberID] = struct{}{}
	}

	entrants, err := s.repo.AllEntrants(ctx)
	if err != nil {
		return err
	}
	candidates := make([]domain.Entrant, 0)
	candidateLogins := make([]string, 0, len(entrants))
	seenLogins := make(map[string]struct{}, len(entrants))
	for _, entrant := range entrants {
		if _, ok := memberSet[fmt.Sprintf("%d", entrant.DiscordID)]; ok {
			login := strings.ToLower(strings.TrimSpace(entrant.GithubLogin))
			if login == "" {
				continue
			}
			candidates = append(candidates, entrant)
			if _, seen := seenLogins[login]; seen {
				continue
			}
			seenLogins[login] = struct{}{}
			candidateLogins = append(candidateLogins, login)
		}
	}

	if len(candidates) == 0 {
		return s.respond(ctx, interaction.Token, "No entrants currently in giveaway role.")
	}

	starred, completed, err := s.github.CheckUsersStar(ctx, candidateLogins, s.owner, s.repoName, s.cfg.WinnerConcurrency)
	if err != nil {
		s.logger.Warn().Err(err).Bool("completed", completed).Msg("star check partially failed")
	}
	if !completed {
		return s.respond(ctx, interaction.Token, "Unable to verify stargazers right now; please retry the winner command.")
	}
	if err != nil {
		return s.respond(ctx, interaction.Token, "Unable to verify stargazers right now; please retry the winner command.")
	}

	kept := make([]domain.Entrant, 0)
	removed := 0
	for _, entrant := range candidates {
		if starred[strings.ToLower(strings.TrimSpace(entrant.GithubLogin))] {
			kept = append(kept, entrant)
			continue
		}
		if err := s.discord.RemoveRoleFromMember(ctx, s.cfg.GuildID, fmt.Sprintf("%d", entrant.DiscordID), s.cfg.GiveawayRoleID); err != nil {
			s.logger.Warn().Err(err).Int64("discord_id", entrant.DiscordID).Msg("remove role failed")
		}
		removed++
	}

	if len(kept) == 0 {
		msg := fmt.Sprintf("No eligible entrants. Removed %d non-starrer(s)", removed)
		return s.respond(ctx, interaction.Token, msg)
	}

	if len(kept) > winnerCount {
		if err := shuffleEntrants(kept); err != nil {
			return s.respond(ctx, interaction.Token, "Unable to draw winners right now; please retry the winner command.")
		}
		kept = kept[:winnerCount]
	}

	lines := []string{"🏆 Winners selected:"}
	for i, winner := range kept {
		lines = append(lines, fmt.Sprintf("%d. <@%d> (%s)", i+1, winner.DiscordID, winner.GithubLogin))
	}
	if removed > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Removed %d entrant(s) who are no longer starred.", removed))
	}
	return s.respond(ctx, interaction.Token, strings.Join(lines, "\n"))
}

func (s *WinnerService) respond(ctx context.Context, interactionToken, content string) error {
	payload := domain.WebhookMessageEdit{
		Content:         content,
		AllowedMentions: &domain.AllowedMentions{Parse: []string{"users"}},
	}
	return s.discord.EditOriginalInteractionResponse(ctx, s.cfg.DiscordApplicationID, interactionToken, payload)
}

func shuffleEntrants(entries []domain.Entrant) error {
	for i := len(entries) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		swap := int(j.Int64())
		entries[i], entries[swap] = entries[swap], entries[i]
	}
	return nil
}
