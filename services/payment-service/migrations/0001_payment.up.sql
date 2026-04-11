-- services/payment-service/migrations/0001_init.up.sql
CREATE TABLE IF NOT EXISTS payments (
    id             UUID PRIMARY KEY,
    appointment_id UUID UNIQUE NOT NULL,
    patient_id     UUID NOT NULL,
    amount         DECIMAL(10,2) NOT NULL,
    currency       VARCHAR(3) NOT NULL,
    status         VARCHAR(20) NOT NULL,
    provider       VARCHAR(20) NOT NULL,
    provider_id    VARCHAR(255),
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_appointment_id ON payments(appointment_id);
CREATE INDEX IF NOT EXISTS idx_payments_patient_id ON payments(patient_id);
