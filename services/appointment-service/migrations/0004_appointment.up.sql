CREATE TABLE IF NOT EXISTS appointments (
    id               UUID PRIMARY KEY,
    patient_id       UUID NOT NULL,
    doctor_id        TEXT NOT NULL,
    scheduled_at     TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 30,
    status           TEXT NOT NULL DEFAULT 'pending',
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS slots (
    id         UUID PRIMARY KEY,
    doctor_id  TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    is_booked  BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_appointments_patient ON appointments(patient_id);
CREATE INDEX IF NOT EXISTS idx_appointments_doctor  ON appointments(doctor_id);
CREATE INDEX IF NOT EXISTS idx_slots_doctor         ON slots(doctor_id);
CREATE INDEX IF NOT EXISTS idx_slots_start_time     ON slots(start_time);
CREATE INDEX IF NOT EXISTS idx_appointments_status  ON appointments(status);
