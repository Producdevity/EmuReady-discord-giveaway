CREATE SEQUENCE IF NOT EXISTS entries_id_seq;

ALTER TABLE entries ADD COLUMN IF NOT EXISTS id BIGINT;
UPDATE entries SET id = nextval('entries_id_seq') WHERE id IS NULL;
SELECT setval('entries_id_seq', COALESCE((SELECT MAX(id) FROM entries), 0) + 1, false);
ALTER TABLE entries ALTER COLUMN id SET DEFAULT nextval('entries_id_seq');
ALTER TABLE entries ALTER COLUMN id SET NOT NULL;
ALTER SEQUENCE entries_id_seq OWNED BY entries.id;

ALTER TABLE entries ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_pkey;
ALTER TABLE entries ADD CONSTRAINT entries_pkey PRIMARY KEY (id);

DROP INDEX IF EXISTS idx_entries_github_id_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_active_discord_id_unique ON entries(discord_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_active_github_id_unique ON entries(github_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_entries_deleted_at ON entries(deleted_at);
