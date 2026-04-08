CREATE TABLE IF NOT EXISTS doctors (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    specialization  VARCHAR(255) NOT NULL,
    experience      INT NOT NULL CHECK (experience >= 0 AND experience <= 80),
    hospital        VARCHAR(255) NOT NULL,
    nic             VARCHAR(12) NOT NULL,
    slmc_no         VARCHAR(5) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_doctors_specialization ON doctors (specialization);
-- Unique indexes on nic/slmc_no are created in 0002 (after columns exist on upgraded DBs).
