ALTER TABLE doctors ADD COLUMN IF NOT EXISTS user_id TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_user_id ON doctors (user_id) WHERE user_id <> '';
