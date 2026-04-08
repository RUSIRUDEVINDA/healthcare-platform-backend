package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"healthcare-platform/services/doctor-service/internal/model"
)

// DoctorRepository handles PostgreSQL access for doctors.
type DoctorRepository struct {
	db *sql.DB
}

func NewDoctorRepository(db *sql.DB) *DoctorRepository {
	return &DoctorRepository{db: db}
}

// Create inserts a doctor and returns the persisted row (including generated id and timestamps).
func (r *DoctorRepository) Create(d *model.Doctor) error {
	query := `
		INSERT INTO doctors (name, specialization, experience, hospital, nic, slmc_no, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	now := time.Now().UTC()
	err := r.db.QueryRow(query,
		d.Name,
		d.Specialization,
		d.Experience,
		d.Hospital,
		d.NIC,
		d.SLMCNo,
		now,
		now,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

// List returns all doctors, optionally filtered by specialization (case-insensitive partial match).
func (r *DoctorRepository) List(specializationFilter string) ([]model.Doctor, error) {
	base := `SELECT id, name, specialization, experience, hospital, nic, slmc_no, created_at, updated_at FROM doctors`
	var args []interface{}
	var sb strings.Builder
	sb.WriteString(base)
	if strings.TrimSpace(specializationFilter) != "" {
		sb.WriteString(` WHERE LOWER(specialization) LIKE LOWER($1)`)
		args = append(args, "%"+strings.TrimSpace(specializationFilter)+"%")
	}
	sb.WriteString(` ORDER BY id ASC`)

	rows, err := r.db.Query(sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}
	defer rows.Close()

	var out []model.Doctor
	for rows.Next() {
		var d model.Doctor
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Specialization, &d.Experience, &d.Hospital,
			&d.NIC, &d.SLMCNo, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository.List scan: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.List rows: %w", err)
	}
	return out, nil
}

// OtherDoctorIDWithNIC returns another doctor's id that already has this NIC, or (0, false).
func (r *DoctorRepository) OtherDoctorIDWithNIC(excludeID int64, nic string) (int64, bool, error) {
	var other int64
	err := r.db.QueryRow(
		`SELECT id FROM doctors WHERE id <> $1 AND nic = $2 LIMIT 1`,
		excludeID, nic,
	).Scan(&other)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("repository.OtherDoctorIDWithNIC: %w", err)
	}
	return other, true, nil
}

// OtherDoctorIDWithSLMC returns another doctor's id that already has this SLMC number, or (0, false).
func (r *DoctorRepository) OtherDoctorIDWithSLMC(excludeID int64, slmcNo string) (int64, bool, error) {
	var other int64
	err := r.db.QueryRow(
		`SELECT id FROM doctors WHERE id <> $1 AND slmc_no = $2 LIMIT 1`,
		excludeID, slmcNo,
	).Scan(&other)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("repository.OtherDoctorIDWithSLMC: %w", err)
	}
	return other, true, nil
}

// FindByID returns a doctor by primary key, or nil if not found.
func (r *DoctorRepository) FindByID(id int64) (*model.Doctor, error) {
	query := `
		SELECT id, name, specialization, experience, hospital, nic, slmc_no, created_at, updated_at
		FROM doctors WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	d := &model.Doctor{}
	err := row.Scan(
		&d.ID, &d.Name, &d.Specialization, &d.Experience, &d.Hospital,
		&d.NIC, &d.SLMCNo, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}
	return d, nil
}

// Update replaces mutable fields for an existing doctor.
func (r *DoctorRepository) Update(d *model.Doctor) error {
	query := `
		UPDATE doctors
		SET name = $1, specialization = $2, experience = $3, hospital = $4,
		    nic = $5, slmc_no = $6, updated_at = $7
		WHERE id = $8
		RETURNING updated_at
	`
	now := time.Now().UTC()
	err := r.db.QueryRow(query,
		d.Name, d.Specialization, d.Experience, d.Hospital, d.NIC, d.SLMCNo, now, d.ID,
	).Scan(&d.UpdatedAt)
	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	}
	if err != nil {
		return fmt.Errorf("repository.Update: %w", err)
	}
	return nil
}

// Delete removes a doctor row by id. Returns rows affected count via caller checking.
func (r *DoctorRepository) Delete(id int64) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM doctors WHERE id = $1`, id)
	if err != nil {
		return 0, fmt.Errorf("repository.Delete: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}
