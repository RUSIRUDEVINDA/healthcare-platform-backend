CREATE TABLE IF NOT EXISTS symptom_chat_history (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             TEXT NOT NULL,
    symptoms            TEXT NOT NULL,
    optional_context    TEXT NOT NULL DEFAULT '',
    suggested_specialty TEXT NOT NULL,
    preliminary_notes   TEXT NOT NULL,
    disclaimer          TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_symptom_chat_user_created ON symptom_chat_history (user_id, created_at DESC);
