CREATE TABLE IF NOT EXISTS appointments (
    id               UUID PRIMARY KEY,
    patient_id       UUID NOT NULL,
    doctor_id        TEXT NOT NULL,
    doctor_owner_user_id TEXT NOT NULL DEFAULT '',
    slot_id          UUID,
    consultation_mode TEXT NOT NULL DEFAULT 'physical',
    room_name        TEXT NOT NULL DEFAULT '',
    join_url         TEXT NOT NULL DEFAULT '',
    scheduled_at     TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 30,
    status           TEXT NOT NULL DEFAULT 'pending',
    payment_status   TEXT NOT NULL DEFAULT 'pending',
    payment_due_at   TIMESTAMPTZ,
    paid_at          TIMESTAMPTZ,
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
CREATE INDEX IF NOT EXISTS idx_appointments_doctor_owner ON appointments(doctor_owner_user_id);
CREATE INDEX IF NOT EXISTS idx_appointments_slot ON appointments(slot_id);
CREATE INDEX IF NOT EXISTS idx_slots_doctor         ON slots(doctor_id);
CREATE INDEX IF NOT EXISTS idx_slots_start_time     ON slots(start_time);
CREATE INDEX IF NOT EXISTS idx_appointments_status  ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_payment_due ON appointments(payment_due_at);
CREATE INDEX IF NOT EXISTS idx_appointments_payment_status ON appointments(payment_status);
