package domain

import "context"

type EntrantStore interface {
	FindEntrant(ctx context.Context, discordID int64) (*Entrant, error)
	UpsertEntrant(ctx context.Context, entrant Entrant) error
	AllEntrants(ctx context.Context) ([]Entrant, error)
	FilterEntrantsByDiscordIDs(ctx context.Context, ids []int64) ([]Entrant, error)
}

type DiscordIntegration interface {
	GetMembersByRole(ctx context.Context, guildID string, roleID string) ([]string, error)
	AddRoleToMember(ctx context.Context, guildID, userID, roleID string) error
	RemoveRoleFromMember(ctx context.Context, guildID, userID, roleID string) error
	EditOriginalInteractionResponse(ctx context.Context, applicationID string, interactionToken string, body interface{}) error
	GetApplicationID(ctx context.Context) (string, error)
	RegisterGuildCommands(ctx context.Context, applicationID string, guildID string, commands interface{}) error
}

type GitHubIntegration interface {
	ExchangeCode(ctx context.Context, code string, redirectURI string) (string, error)
	GetCurrentUser(ctx context.Context, accessToken string) (GitHubUser, error)
	HasStarredRepo(ctx context.Context, accessToken string, owner, repo string) (bool, error)
	CheckUsersStar(ctx context.Context, usernames []string, owner, repo string, maxPages int) (map[string]bool, bool, error)
}

type GitHubUser struct {
	ID    int64
	Login string
}
