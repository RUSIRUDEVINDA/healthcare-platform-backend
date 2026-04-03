CREATE TABLE IF NOT EXISTS patients (
	id                 UUID PRIMARY KEY,
	user_id            UUID UNIQUE NOT NULL,
	email              VARCHAR(255) NOT NULL,
	first_name         VARCHAR(100) NOT NULL,
	last_name          VARCHAR(100) NOT NULL,
	date_of_birth      DATE,
	gender             VARCHAR(20),
	phone_number       VARCHAR(20),
	address            TEXT,
	emergency_contact  VARCHAR(255),
	blood_group        VARCHAR(5),
	created_at         TIMESTAMPTZ DEFAULT NOW(),
	updated_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patients_user_id ON patients(user_id);
CREATE INDEX IF NOT EXISTS idx_patients_email ON patients(email);
