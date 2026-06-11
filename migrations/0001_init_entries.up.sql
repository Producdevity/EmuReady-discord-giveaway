CREATE TABLE IF NOT EXISTS entries (
  discord_id BIGINT PRIMARY KEY,
  github_id BIGINT NOT NULL,
  github_login TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entries_created_at ON entries(created_at);
