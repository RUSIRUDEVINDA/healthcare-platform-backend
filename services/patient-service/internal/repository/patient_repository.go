package repository

import (
	"database/sql"
	"fmt"

	"healthcare-platform/services/patient-service/internal/model"
	"time"
)

type PatientRepository struct {
	db *sql.DB
}

func NewPatientRepository(db *sql.DB) *PatientRepository {
	return &PatientRepository{db: db}
}

func (r *PatientRepository) Create(p *model.Patient) error {
	query := `
		INSERT INTO patients (id, user_id, email, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE
		SET email = EXCLUDED.email,
		    first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, p.ID, p.UserID, p.Email, p.FirstName, p.LastName, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

func (r *PatientRepository) FindByUserID(userID string) (*model.Patient, error) {
	query := `
		SELECT id, user_id, email, first_name, last_name, date_of_birth, gender, phone_number, address, emergency_contact, blood_group, created_at, updated_at
		FROM patients
		WHERE user_id = $1
	`
	row := r.db.QueryRow(query, userID)
	var p model.Patient
	err := row.Scan(
		&p.ID, &p.UserID, &p.Email, &p.FirstName, &p.LastName, &p.DateOfBirth, &p.Gender, &p.PhoneNumber, &p.Address, &p.EmergencyContact, &p.BloodGroup, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByUserID: %w", err)
	}
	return &p, nil
}

func (r *PatientRepository) Update(userID string, p *model.UpdatePatientRequest) error {
	dateOfBirth := flexibleDateToTime(p.DateOfBirth)

	query := `
		UPDATE patients
		SET date_of_birth = $1, gender = $2, phone_number = $3, address = $4, emergency_contact = $5, blood_group = $6, updated_at = NOW()
		WHERE user_id = $7
	`
	_, err := r.db.Exec(query, dateOfBirth, p.Gender, p.PhoneNumber, p.Address, p.EmergencyContact, p.BloodGroup, userID)
	if err != nil {
		return fmt.Errorf("repository.Update: %w", err)
	}
	return nil
}

func (r *PatientRepository) UpdatePartial(userID string, p *model.PatchPatientRequest) error {
	dateOfBirth := flexibleDateToTime(p.DateOfBirth)

	query := `
		UPDATE patients
		SET date_of_birth = COALESCE($1, date_of_birth),
		    gender = COALESCE($2, gender),
		    phone_number = COALESCE($3, phone_number),
		    address = COALESCE($4, address),
		    emergency_contact = COALESCE($5, emergency_contact),
		    blood_group = COALESCE($6, blood_group),
		    updated_at = NOW()
		WHERE user_id = $7
	`
	_, err := r.db.Exec(query, dateOfBirth, p.Gender, p.PhoneNumber, p.Address, p.EmergencyContact, p.BloodGroup, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdatePartial: %w", err)
	}
	return nil
}

func (r *PatientRepository) DeleteByUserID(userID string) (bool, error) {
	query := `DELETE FROM patients WHERE user_id = $1`
	res, err := r.db.Exec(query, userID)
	if err != nil {
		return false, fmt.Errorf("repository.DeleteByUserID: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("repository.DeleteByUserID rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func flexibleDateToTime(value *model.FlexibleTime) *time.Time {
	if value == nil {
		return nil
	}
	t := value.Time
	return &t
}
