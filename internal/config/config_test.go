package config

import "testing"

func TestLoadAcceptsValidConfig(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("base URL mismatch: %q", cfg.BaseURL)
	}
	if cfg.WinnerDefaultCount != cfg.WinnerMax {
		t.Fatalf("expected default winner count to clamp to max, got default=%d max=%d", cfg.WinnerDefaultCount, cfg.WinnerMax)
	}
}

func TestLoadRejectsPlaceholdersAndInvalidDiscordPublicKey(t *testing.T) {
	setValidEnv(t)
	t.Setenv("SIGNING_SECRET", "__REQUIRED__")
	if _, err := Load(); err == nil {
		t.Fatal("expected placeholder signing secret to fail")
	}

	setValidEnv(t)
	t.Setenv("DISCORD_PUBLIC_KEY", "abc")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid public key to fail")
	}
}

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DISCORD_TOKEN", "discord-token")
	t.Setenv("DISCORD_PUBLIC_KEY", "0000000000000000000000000000000000000000000000000000000000000000")
	t.Setenv("GUILD_ID", "123456789012345678")
	t.Setenv("GIVEAWAY_ROLE_ID", "123456789012345679")
	t.Setenv("GITHUB_CLIENT_ID", "github-client")
	t.Setenv("GITHUB_CLIENT_SECRET", "github-secret")
	t.Setenv("GITHUB_REPO", "owner/repo")
	t.Setenv("BASE_URL", "https://example.com/")
	t.Setenv("SIGNING_SECRET", "this-secret-is-long-enough-for-tests")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/giveaway?sslmode=disable")
	t.Setenv("WINNER_DEFAULT_COUNT", "10")
	t.Setenv("WINNER_MAX", "5")
}
