ALTER TABLE album ALTER COLUMN id TYPE VARCHAR(36);
ALTER TABLE album ALTER COLUMN name TYPE VARCHAR(128);

-- Default timestamps so the DB supplies the value when the app omits them.
ALTER TABLE album ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE album ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- Enforce uniqueness of album names at the schema level.
ALTER TABLE album ADD CONSTRAINT album_name_unique UNIQUE (name);

-- Prevent updated_at from being set to a time before created_at.
ALTER TABLE album ADD CONSTRAINT album_timestamps_check CHECK (updated_at >= created_at);

CREATE INDEX idx_album_name       ON album (name);
CREATE INDEX idx_album_created_at ON album (created_at);
CREATE INDEX idx_album_updated_at ON album (updated_at);
