package storage

import "context"

import "github.com/producdevity/emuready-discord-giveaway/internal/domain"

type EntrantRepository interface {
	FindEntrant(ctx context.Context, discordID int64) (*domain.Entrant, error)
	UpsertEntrant(ctx context.Context, entrant domain.Entrant) error
	DeleteEntrant(ctx context.Context, discordID int64, githubID int64) error
	CountEntrants(ctx context.Context) (int, error)
	AllEntrants(ctx context.Context) ([]domain.Entrant, error)
}
