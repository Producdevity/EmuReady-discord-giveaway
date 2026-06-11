package config

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Environment string `env:"ENV" envDefault:"production"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	Port        string `env:"PORT" envDefault:"8080"`

	DiscordToken         string `env:"DISCORD_TOKEN,required"`
	DiscordPublicKey     string `env:"DISCORD_PUBLIC_KEY,required"`
	DiscordApplicationID string `env:"DISCORD_APPLICATION_ID"`
	GuildID              string `env:"GUILD_ID,required"`
	GiveawayRoleID       string `env:"GIVEAWAY_ROLE_ID,required"`

	GithubClientID     string `env:"GITHUB_CLIENT_ID,required"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET,required"`
	GithubRepo         string `env:"GITHUB_REPO,required"`
	GithubOAuthScope   string `env:"GITHUB_OAUTH_SCOPE"`
	BaseURL            string `env:"BASE_URL,required"`
	SigningSecret      string `env:"SIGNING_SECRET,required"`
	DatabaseURL        string `env:"DATABASE_URL,required"`
	GithubApiToken     string `env:"GITHUB_API_TOKEN"`
	MigrationsDir      string `env:"MIGRATIONS_DIR" envDefault:"migrations"`

	StateTTLSecs            int `env:"STATE_TTL_SECONDS" envDefault:"600"`
	WinnerDefaultCount      int `env:"WINNER_DEFAULT_COUNT" envDefault:"1"`
	WinnerMax               int `env:"WINNER_MAX" envDefault:"50"`
	WinnerConcurrency       int `env:"WINNER_CONCURRENCY" envDefault:"10"`
	WinnerWorkerCount       int `env:"WINNER_WORKER_COUNT" envDefault:"2"`
	HTTPTimeoutSeconds      int `env:"HTTP_TIMEOUT_SECONDS" envDefault:"10"`
	DBConnectTimeoutSeconds int `env:"DB_CONNECT_TIMEOUT_SECONDS" envDefault:"30"`
	MaxInteractionBodyBytes int `env:"MAX_INTERACTION_BODY_BYTES" envDefault:"65536"`
	HTTPMaxRetries          int `env:"HTTP_MAX_RETRIES" envDefault:"3"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	required := map[string]string{
		"DISCORD_TOKEN":        cfg.DiscordToken,
		"DISCORD_PUBLIC_KEY":   cfg.DiscordPublicKey,
		"GUILD_ID":             cfg.GuildID,
		"GIVEAWAY_ROLE_ID":     cfg.GiveawayRoleID,
		"GITHUB_CLIENT_ID":     cfg.GithubClientID,
		"GITHUB_CLIENT_SECRET": cfg.GithubClientSecret,
		"GITHUB_REPO":          cfg.GithubRepo,
		"BASE_URL":             cfg.BaseURL,
		"SIGNING_SECRET":       cfg.SigningSecret,
		"DATABASE_URL":         cfg.DatabaseURL,
	}
	for name, value := range required {
		if isPlaceholder(value) {
			return nil, fmt.Errorf("%s must be configured", name)
		}
	}
	if len(cfg.SigningSecret) < 32 {
		return nil, fmt.Errorf("SIGNING_SECRET must be at least 32 characters")
	}
	key, err := hex.DecodeString(strings.TrimSpace(cfg.DiscordPublicKey))
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("DISCORD_PUBLIC_KEY must be a 32-byte hex-encoded Ed25519 public key")
	}
	if _, _, err := cfg.GitHubRepoParts(); err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("BASE_URL must be an absolute URL")
	}
	cfg.BaseURL = strings.TrimRight(baseURL.String(), "/")
	cfg.MigrationsDir = strings.TrimSpace(cfg.MigrationsDir)
	if cfg.MigrationsDir == "" {
		cfg.MigrationsDir = "migrations"
	}
	if port, err := strconv.Atoi(strings.TrimSpace(cfg.Port)); err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("PORT must be a TCP port between 1 and 65535")
	}
	if cfg.StateTTLSecs <= 0 {
		cfg.StateTTLSecs = 600
	}
	if cfg.WinnerDefaultCount <= 0 {
		cfg.WinnerDefaultCount = 1
	}
	if cfg.WinnerMax <= 0 {
		cfg.WinnerMax = 50
	}
	if cfg.WinnerDefaultCount > cfg.WinnerMax {
		cfg.WinnerDefaultCount = cfg.WinnerMax
	}
	if cfg.WinnerConcurrency <= 0 {
		cfg.WinnerConcurrency = 10
	}
	if cfg.WinnerWorkerCount <= 0 {
		cfg.WinnerWorkerCount = 2
	}
	if cfg.HTTPTimeoutSeconds <= 0 {
		cfg.HTTPTimeoutSeconds = 10
	}
	if cfg.DBConnectTimeoutSeconds <= 0 {
		cfg.DBConnectTimeoutSeconds = 30
	}
	if cfg.MaxInteractionBodyBytes <= 0 {
		cfg.MaxInteractionBodyBytes = 64 * 1024
	}
	if cfg.MaxInteractionBodyBytes > 10*1024*1024 {
		cfg.MaxInteractionBodyBytes = 10 * 1024 * 1024
	}
	if cfg.HTTPMaxRetries <= 0 {
		cfg.HTTPMaxRetries = 3
	}
	return cfg, nil
}

func (c *Config) StateTTL() time.Duration {
	return time.Duration(c.StateTTLSecs) * time.Second
}

func (c *Config) HTTPTimeout() time.Duration {
	return time.Duration(c.HTTPTimeoutSeconds) * time.Second
}

func (c *Config) DBConnectTimeout() time.Duration {
	return time.Duration(c.DBConnectTimeoutSeconds) * time.Second
}

func (c *Config) GitHubRepoParts() (owner string, repo string, err error) {
	parts := strings.Split(strings.TrimSpace(c.GithubRepo), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("GITHUB_REPO must be in owner/repo format")
	}
	return parts[0], parts[1], nil
}

func isPlaceholder(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == "__REQUIRED__"
}
