CREATE TABLE IF NOT EXISTS telemedicine_sessions (
	id UUID PRIMARY KEY,
	appointment_id UUID NOT NULL,
	session_number INT NOT NULL,
	patient_id UUID NOT NULL,
	doctor_id TEXT NOT NULL,
	doctor_owner_user_id TEXT NOT NULL DEFAULT '',
	room_name TEXT NOT NULL UNIQUE,
	join_url TEXT NOT NULL,
	provider TEXT NOT NULL DEFAULT 'jitsi',
	status TEXT NOT NULL DEFAULT 'scheduled',
	purpose TEXT,
	notes TEXT,
	created_by_user_id TEXT NOT NULL,
	created_by_role TEXT NOT NULL,
	started_at TIMESTAMPTZ,
	ended_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT uq_telemedicine_sessions_appointment_sequence UNIQUE (appointment_id, session_number)
);

CREATE INDEX IF NOT EXISTS idx_telemedicine_sessions_appointment ON telemedicine_sessions(appointment_id);
CREATE INDEX IF NOT EXISTS idx_telemedicine_sessions_patient ON telemedicine_sessions(patient_id);
CREATE INDEX IF NOT EXISTS idx_telemedicine_sessions_doctor_owner ON telemedicine_sessions(doctor_owner_user_id);
CREATE INDEX IF NOT EXISTS idx_telemedicine_sessions_room_name ON telemedicine_sessions(room_name);
CREATE INDEX IF NOT EXISTS idx_telemedicine_sessions_status ON telemedicine_sessions(status);

