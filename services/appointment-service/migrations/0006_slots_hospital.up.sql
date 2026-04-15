ALTER TABLE slots
    ADD COLUMN IF NOT EXISTS hospital TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_slots_hospital ON slots(hospital);
