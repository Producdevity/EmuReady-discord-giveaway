package application

import (
	"net/url"
	"strings"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
)

type EnterService struct {
	cfg   *config.Config
	state *OAuthStateService
}

func NewEnterService(cfg *config.Config, state *OAuthStateService) *EnterService {
	return &EnterService{cfg: cfg, state: state}
}

func (s *EnterService) BuildResponse(discordID int64) (domain.InteractionResponse, error) {
	state, err := s.state.CreateState(discordID)
	if err != nil {
		return domain.InteractionResponse{}, err
	}
	owner, repo, err := s.cfg.GitHubRepoParts()
	if err != nil {
		return domain.InteractionResponse{}, err
	}

	callbackURL := strings.TrimRight(s.cfg.BaseURL, "/") + "/callback"
	values := url.Values{
		"client_id":    []string{s.cfg.GithubClientID},
		"redirect_uri": []string{callbackURL},
		"state":        []string{state},
	}
	if scope := strings.TrimSpace(s.cfg.GithubOAuthScope); scope != "" {
		values.Set("scope", scope)
	}
	authURL := "https://github.com/login/oauth/authorize?" + values.Encode()
	repoName := owner + "/" + repo
	repoURL := "https://github.com/" + repoName

	return domain.InteractionResponse{
		Type: domain.InteractionResponseChannelMessage,
		Data: &domain.InteractionMessageData{
			Content: "Star " + repoName + ", then authenticate with GitHub to enter the giveaway.",
			Flags:   domain.MessageFlagEphemeral,
			Components: []any{
				map[string]any{
					"type": 1,
					"components": []any{
						map[string]any{
							"type":     2,
							"style":    5,
							"label":    "Open GitHub repo",
							"url":      repoURL,
							"disabled": false,
						},
						map[string]any{
							"type":     2,
							"style":    5,
							"label":    "Authenticate on GitHub",
							"url":      authURL,
							"disabled": false,
						},
					},
				},
			},
		},
	}, nil
}
