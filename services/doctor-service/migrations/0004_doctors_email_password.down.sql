DROP INDEX IF EXISTS idx_doctors_email;
ALTER TABLE doctors DROP COLUMN IF EXISTS password_hash;
ALTER TABLE doctors DROP COLUMN IF EXISTS email;
