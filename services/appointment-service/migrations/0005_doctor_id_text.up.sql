-- Allow doctor_id to be either a UUID or numeric string during transition period.
-- Older DBs used UUID type for doctor_id, which rejects numeric values.

ALTER TABLE appointments
    ALTER COLUMN doctor_id TYPE TEXT
    USING doctor_id::text;

ALTER TABLE slots
    ALTER COLUMN doctor_id TYPE TEXT
    USING doctor_id::text;