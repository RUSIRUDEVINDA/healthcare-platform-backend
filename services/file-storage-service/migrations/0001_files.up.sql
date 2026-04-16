CREATE TABLE IF NOT EXISTS files (
	id UUID PRIMARY KEY,
	owner_id TEXT NOT NULL,
	uploader_id TEXT NOT NULL,
	kind TEXT NOT NULL,
	storage_provider TEXT NOT NULL,
	original_name TEXT NOT NULL,
	stored_name TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	size_bytes BIGINT NOT NULL DEFAULT 0,
	checksum TEXT NOT NULL DEFAULT '',
	cloudinary_public_id TEXT NOT NULL DEFAULT '',
	cloudinary_url TEXT NOT NULL DEFAULT '',
	r2_bucket TEXT NOT NULL DEFAULT '',
	r2_object_key TEXT NOT NULL DEFAULT '',
	is_public BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

ALTER TABLE files ADD COLUMN IF NOT EXISTS r2_bucket TEXT NOT NULL DEFAULT '';
ALTER TABLE files ADD COLUMN IF NOT EXISTS r2_object_key TEXT NOT NULL DEFAULT '';
ALTER TABLE files DROP COLUMN IF EXISTS gcs_bucket;
ALTER TABLE files DROP COLUMN IF EXISTS gcs_object_name;

CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files(owner_id);
CREATE INDEX IF NOT EXISTS idx_files_uploader_id ON files(uploader_id);
CREATE INDEX IF NOT EXISTS idx_files_kind ON files(kind);
CREATE INDEX IF NOT EXISTS idx_files_storage_provider ON files(storage_provider);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
