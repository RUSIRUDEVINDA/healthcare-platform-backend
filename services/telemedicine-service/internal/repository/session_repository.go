package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"healthcare-platform/services/telemedicine-service/internal/model"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(s *model.Session) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var nextSessionNumber int
	if err := tx.QueryRow(`
		SELECT COALESCE(MAX(session_number), 0) + 1
		FROM telemedicine_sessions
		WHERE appointment_id = $1`,
		s.AppointmentID,
	).Scan(&nextSessionNumber); err != nil {
		return err
	}
	s.SessionNumber = nextSessionNumber

	_, err = tx.Exec(`
		INSERT INTO telemedicine_sessions (
			id, appointment_id, session_number, patient_id, doctor_id, doctor_owner_user_id,
			room_name, join_url, provider, status, purpose, notes,
			created_by_user_id, created_by_role, started_at, ended_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		s.ID, s.AppointmentID, s.SessionNumber, s.PatientID, s.DoctorID, s.DoctorOwnerUserID,
		s.RoomName, s.JoinURL, s.Provider, s.Status, s.Purpose, s.Notes,
		s.CreatedByUserID, s.CreatedByRole, s.StartedAt, s.EndedAt, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SessionRepository) GetByID(id string) (*model.Session, error) {
	s := &model.Session{}
	var startedAt sql.NullTime
	var endedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, appointment_id, session_number, patient_id, doctor_id, doctor_owner_user_id,
		       room_name, join_url, provider, status, purpose, notes,
		       created_by_user_id, created_by_role, started_at, ended_at, created_at, updated_at
		FROM telemedicine_sessions
		WHERE id = $1`, id,
	).Scan(
		&s.ID, &s.AppointmentID, &s.SessionNumber, &s.PatientID, &s.DoctorID, &s.DoctorOwnerUserID,
		&s.RoomName, &s.JoinURL, &s.Provider, &s.Status, &s.Purpose, &s.Notes,
		&s.CreatedByUserID, &s.CreatedByRole, &startedAt, &endedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t := startedAt.Time
		s.StartedAt = &t
	}
	if endedAt.Valid {
		t := endedAt.Time
		s.EndedAt = &t
	}
	return s, nil
}

func (r *SessionRepository) FindByRoomName(roomName string) (*model.Session, error) {
	s := &model.Session{}
	var startedAt sql.NullTime
	var endedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, appointment_id, session_number, patient_id, doctor_id, doctor_owner_user_id,
		       room_name, join_url, provider, status, purpose, notes,
		       created_by_user_id, created_by_role, started_at, ended_at, created_at, updated_at
		FROM telemedicine_sessions
		WHERE room_name = $1`, roomName,
	).Scan(
		&s.ID, &s.AppointmentID, &s.SessionNumber, &s.PatientID, &s.DoctorID, &s.DoctorOwnerUserID,
		&s.RoomName, &s.JoinURL, &s.Provider, &s.Status, &s.Purpose, &s.Notes,
		&s.CreatedByUserID, &s.CreatedByRole, &startedAt, &endedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t := startedAt.Time
		s.StartedAt = &t
	}
	if endedAt.Valid {
		t := endedAt.Time
		s.EndedAt = &t
	}
	return s, nil
}

func (r *SessionRepository) ListByAppointment(appointmentID string) ([]model.Session, error) {
	rows, err := r.db.Query(`
		SELECT id, appointment_id, session_number, patient_id, doctor_id, doctor_owner_user_id,
		       room_name, join_url, provider, status, purpose, notes,
		       created_by_user_id, created_by_role, started_at, ended_at, created_at, updated_at
		FROM telemedicine_sessions
		WHERE appointment_id = $1
		ORDER BY session_number DESC, created_at DESC`,
		appointmentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var startedAt sql.NullTime
		var endedAt sql.NullTime
		if err := rows.Scan(
			&s.ID, &s.AppointmentID, &s.SessionNumber, &s.PatientID, &s.DoctorID, &s.DoctorOwnerUserID,
			&s.RoomName, &s.JoinURL, &s.Provider, &s.Status, &s.Purpose, &s.Notes,
			&s.CreatedByUserID, &s.CreatedByRole, &startedAt, &endedAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			t := startedAt.Time
			s.StartedAt = &t
		}
		if endedAt.Valid {
			t := endedAt.Time
			s.EndedAt = &t
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *SessionRepository) MarkStarted(id string) error {
	res, err := r.db.Exec(`
		UPDATE telemedicine_sessions
		SET status = $1,
		    started_at = COALESCE(started_at, NOW()),
		    updated_at = NOW()
		WHERE id = $2
		  AND status IN ($3, $4)`,
		model.SessionActive, id, model.SessionScheduled, model.SessionActive,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found or cannot be started")
	}
	return nil
}

func (r *SessionRepository) MarkEnded(id string) error {
	res, err := r.db.Exec(`
		UPDATE telemedicine_sessions
		SET status = $1,
		    ended_at = COALESCE(ended_at, NOW()),
		    updated_at = NOW()
		WHERE id = $2
		  AND status <> $3`,
		model.SessionEnded, id, model.SessionCancelled,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found or cannot be ended")
	}
	return nil
}

func (r *SessionRepository) DeleteByPatientID(patientID string) error {
	_, err := r.db.Exec(`DELETE FROM telemedicine_sessions WHERE patient_id = $1`, patientID)
	if err != nil {
		return fmt.Errorf("repository.DeleteByPatientID: %w", err)
	}
	return nil
}
