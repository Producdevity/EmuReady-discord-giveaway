package main

import (
	"context"
	"fmt"

	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/httpclient"
	"github.com/producdevity/emuready-discord-giveaway/internal/integrations/discord"
	"github.com/producdevity/emuready-discord-giveaway/internal/observability"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, err := observability.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	httpClient := httpclient.NewHTTPClient(cfg.HTTPTimeout())
	retryPolicy := httpclient.DefaultRetryPolicy()
	retryPolicy.MaxAttempts = cfg.HTTPMaxRetries
	client := discord.NewClient(httpClient, cfg.DiscordToken, logger)
	client.SetRetryPolicy(retryPolicy)

	appID := cfg.DiscordApplicationID
	if appID == "" {
		appID, err = client.GetApplicationID(context.Background())
		if err != nil {
			panic(err)
		}
	}

	commands := []map[string]any{
		{
			"name":        "enter",
			"description": "Enter the giveaway",
			"type":        1,
		},
		{
			"name":        "enter-giveaway",
			"description": "Enter the giveaway",
			"type":        1,
		},
		{
			"name":                       "entrants",
			"description":                "Count entrants",
			"type":                       1,
			"default_member_permissions": fmt.Sprintf("%d", int64(domain.PermissionManageGuild)),
		},
		{
			"name":        "winner",
			"description": "Draw giveaway winners",
			"type":        1,
			"options": []map[string]any{
				{
					"type":        4,
					"name":        "count",
					"description": "Number of winners",
					"required":    false,
					"min_value":   1,
					"max_value":   float64(cfg.WinnerMax),
				},
			},
			"default_member_permissions": fmt.Sprintf("%d", int64(domain.PermissionManageGuild)),
		},
		{
			"name":                       "reset-giveaway",
			"description":                "Archive current giveaway entries",
			"type":                       1,
			"default_member_permissions": fmt.Sprintf("%d", int64(domain.PermissionManageGuild)),
			"options": []map[string]any{
				{
					"type":        3,
					"name":        "confirm",
					"description": "Type RESET to confirm",
					"required":    true,
				},
			},
		},
	}

	if err := client.RegisterGuildCommands(context.Background(), appID, cfg.GuildID, commands); err != nil {
		panic(err)
	}
	log.Info().Str("application", appID).Str("guild", cfg.GuildID).Msg("commands registered")
}
