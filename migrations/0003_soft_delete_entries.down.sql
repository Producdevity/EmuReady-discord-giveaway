DELETE FROM entries WHERE deleted_at IS NOT NULL;

DROP INDEX IF EXISTS idx_entries_deleted_at;
DROP INDEX IF EXISTS idx_entries_active_github_id_unique;
DROP INDEX IF EXISTS idx_entries_active_discord_id_unique;

ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_pkey;
ALTER TABLE entries ADD CONSTRAINT entries_pkey PRIMARY KEY (discord_id);

ALTER TABLE entries ALTER COLUMN id DROP DEFAULT;
ALTER TABLE entries DROP COLUMN IF EXISTS id;
DROP SEQUENCE IF EXISTS entries_id_seq;

ALTER TABLE entries DROP COLUMN IF EXISTS deleted_at;

CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_github_id_unique ON entries(github_id);
