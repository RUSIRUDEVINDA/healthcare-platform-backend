-- Older images/schemas used UUID for user_id; comparing UUID to '' in the index predicate fails (22P02).
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.table_name = 'doctors'
          AND c.column_name = 'user_id'
          AND c.data_type = 'uuid'
    ) THEN
        DROP INDEX IF EXISTS idx_doctors_user_id;
        ALTER TABLE doctors
            ALTER COLUMN user_id TYPE TEXT USING COALESCE(user_id::text, '');
        ALTER TABLE doctors ALTER COLUMN user_id SET DEFAULT '';
        ALTER TABLE doctors ALTER COLUMN user_id SET NOT NULL;
    END IF;
END $$;

ALTER TABLE doctors ADD COLUMN IF NOT EXISTS user_id TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctors_user_id ON doctors (user_id) WHERE user_id <> '';
