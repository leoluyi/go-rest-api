ALTER TABLE album ALTER COLUMN id TYPE VARCHAR(36);
ALTER TABLE album ALTER COLUMN name TYPE VARCHAR(128);

CREATE INDEX idx_album_created_at ON album (created_at);
CREATE INDEX idx_album_updated_at ON album (updated_at);
