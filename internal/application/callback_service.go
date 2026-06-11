package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage"

	"github.com/rs/zerolog"
)

type callbackDiscordClient interface {
	AddRoleToMember(ctx context.Context, guildID, userID, roleID string) error
}

type callbackGitHubClient interface {
	ExchangeCode(ctx context.Context, code string, redirectURI string) (string, error)
	GetCurrentUser(ctx context.Context, accessToken string) (domain.GitHubUser, error)
	HasStarredRepo(ctx context.Context, accessToken string, owner, repo string) (bool, error)
}

type CallbackService struct {
	cfg     *config.Config
	state   *OAuthStateService
	store   storage.EntrantRepository
	discord callbackDiscordClient
	github  callbackGitHubClient
	owner   string
	repo    string
	logger  zerolog.Logger
}

func NewCallbackService(
	cfg *config.Config,
	state *OAuthStateService,
	store storage.EntrantRepository,
	discordClient callbackDiscordClient,
	githubClient callbackGitHubClient,
	logger zerolog.Logger,
) (*CallbackService, error) {
	owner, repo, err := cfg.GitHubRepoParts()
	if err != nil {
		return nil, err
	}
	return &CallbackService{
		cfg:     cfg,
		state:   state,
		store:   store,
		discord: discordClient,
		github:  githubClient,
		owner:   owner,
		repo:    repo,
		logger:  logger,
	}, nil
}

func (s *CallbackService) Handle(ctx context.Context, code, state string) (domain.GitHubUser, error) {
	discordID, err := s.state.VerifyState(state)
	if err != nil {
		return domain.GitHubUser{}, err
	}

	redirectURI := strings.TrimRight(s.cfg.BaseURL, "/") + "/callback"
	accessToken, err := s.github.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		s.logger.Warn().Err(err).Msg("exchange code failed")
		return domain.GitHubUser{}, err
	}

	user, err := s.github.GetCurrentUser(ctx, accessToken)
	if err != nil {
		s.logger.Warn().Err(err).Msg("github user fetch failed")
		return domain.GitHubUser{}, err
	}
	if user.Login == "" || user.ID == 0 {
		s.logger.Warn().Msg("github user missing identity")
		return domain.GitHubUser{}, fmt.Errorf("invalid github profile")
	}

	hasStar, err := s.github.HasStarredRepo(ctx, accessToken, s.owner, s.repo)
	if err != nil {
		s.logger.Warn().Err(err).Msg("star verification failed")
		return domain.GitHubUser{}, err
	}
	if !hasStar {
		return domain.GitHubUser{}, fmt.Errorf("user does not star %s/%s", s.owner, s.repo)
	}

	if err := s.store.UpsertEntrant(ctx, domain.Entrant{DiscordID: discordID, GithubID: user.ID, GithubLogin: user.Login}); err != nil {
		s.logger.Error().Err(err).Int64("discord_id", discordID).Str("github", user.Login).Msg("upsert failed")
		return domain.GitHubUser{}, err
	}

	if err := s.discord.AddRoleToMember(ctx, s.cfg.GuildID, fmt.Sprintf("%d", discordID), s.cfg.GiveawayRoleID); err != nil {
		s.logger.Error().Err(err).Int64("discord_id", discordID).Msg("failed to add role")
		if cleanupErr := s.store.DeleteEntrant(ctx, discordID, user.ID); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Int64("discord_id", discordID).Msg("failed to roll back entrant after role failure")
		}
		return domain.GitHubUser{}, err
	}

	return user, nil
}
