package postgres

import (
	"context"
	"errors"

	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntrantRepository struct {
	pool *pgxpool.Pool
}

func NewEntrantRepository(pool *pgxpool.Pool) storage.EntrantRepository {
	return &EntrantRepository{pool: pool}
}

func (r *EntrantRepository) FindEntrant(ctx context.Context, discordID int64) (*domain.Entrant, error) {
	row := r.pool.QueryRow(ctx, `
	SELECT discord_id, github_id, github_login, created_at
	FROM entries WHERE discord_id = $1`, discordID)

	var entrant domain.Entrant
	if err := row.Scan(&entrant.DiscordID, &entrant.GithubID, &entrant.GithubLogin, &entrant.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &entrant, nil
}

func (r *EntrantRepository) UpsertEntrant(ctx context.Context, entrant domain.Entrant) error {
	var linkedDiscordID int64
	err := r.pool.QueryRow(ctx, `
		SELECT discord_id
		FROM entries
		WHERE github_id = $1 AND discord_id <> $2`, entrant.GithubID, entrant.DiscordID).Scan(&linkedDiscordID)
	if err == nil {
		return domain.ErrGitHubAlreadyLinked
	}
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO entries (discord_id, github_id, github_login)
		VALUES ($1, $2, $3)
		ON CONFLICT (discord_id)
		DO UPDATE SET github_id = EXCLUDED.github_id, github_login = EXCLUDED.github_login`, entrant.DiscordID, entrant.GithubID, entrant.GithubLogin)
	if isUniqueViolation(err) {
		return domain.ErrGitHubAlreadyLinked
	}
	return err
}

func (r *EntrantRepository) DeleteEntrant(ctx context.Context, discordID int64, githubID int64) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM entries
		WHERE discord_id = $1 AND github_id = $2`, discordID, githubID)
	return err
}

func (r *EntrantRepository) CountEntrants(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM entries`).Scan(&count)
	return count, err
}

func (r *EntrantRepository) AllEntrants(ctx context.Context) ([]domain.Entrant, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT discord_id, github_id, github_login, created_at
	FROM entries ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.Entrant, 0)
	for rows.Next() {
		var entrant domain.Entrant
		if err := rows.Scan(&entrant.DiscordID, &entrant.GithubID, &entrant.GithubLogin, &entrant.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entrant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
