package domain

import "time"

type Entrant struct {
	DiscordID   int64     `json:"discord_id"`
	GithubID    int64     `json:"github_id"`
	GithubLogin string    `json:"github_login"`
	CreatedAt   time.Time `json:"created_at"`
}
