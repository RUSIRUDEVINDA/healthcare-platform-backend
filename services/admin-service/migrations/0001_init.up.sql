CREATE TABLE IF NOT EXISTS admin_users (
	id UUID PRIMARY KEY,
	email VARCHAR(255) UNIQUE NOT NULL,
	role VARCHAR(20) NOT NULL,
	first_name VARCHAR(100) NOT NULL,
	last_name VARCHAR(100) NOT NULL,
	is_verified BOOLEAN DEFAULT FALSE,
	is_active BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS doctor_verifications (
	doctor_id UUID PRIMARY KEY REFERENCES admin_users(id) ON DELETE CASCADE,
	verified_by UUID,
	notes TEXT,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	verified_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS appointments (
	id UUID PRIMARY KEY,
	patient_id UUID NOT NULL,
	doctor_id UUID NOT NULL,
	status VARCHAR(30) NOT NULL,
	scheduled_at TIMESTAMPTZ NOT NULL,
	reason TEXT,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	amount NUMERIC(12,2) NOT NULL,
	currency VARCHAR(10) NOT NULL,
	status VARCHAR(30) NOT NULL,
	provider VARCHAR(50) NOT NULL,
	reference VARCHAR(100) UNIQUE,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_doctor_verifications_status ON doctor_verifications(status);
CREATE INDEX IF NOT EXISTS idx_appointments_doctor_id ON appointments(doctor_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
