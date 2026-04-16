package repository

import (
	"database/sql"
	"fmt"
	"time"

	"healthcare-platform/services/admin-service/internal/model"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) UpsertUser(user *model.User) error {
	query := `
		INSERT INTO admin_users (id, email, role, first_name, last_name, is_verified, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			role = EXCLUDED.role,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			is_verified = EXCLUDED.is_verified,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, user.ID, user.Email, string(user.Role), user.FirstName, user.LastName, user.IsVerified, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository.UpsertUser: %w", err)
	}
	return nil
}

func (r *AdminRepository) ListUsers() ([]model.User, error) {
	rows, err := r.db.Query(`
		SELECT id, email, role, first_name, last_name, is_verified, is_active, created_at, updated_at
		FROM admin_users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("repository.ListUsers query: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var role string
		if err := rows.Scan(&user.ID, &user.Email, &role, &user.FirstName, &user.LastName, &user.IsVerified, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository.ListUsers scan: %w", err)
		}
		user.Role = model.Role(role)
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *AdminRepository) VerifyDoctor(doctorID, verifiedBy, notes string) (*model.DoctorVerification, error) {
	now := time.Now().UTC()
	query := `
		INSERT INTO doctor_verifications (doctor_id, verified_by, notes, status, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'verified', $4, $4, $4)
		ON CONFLICT (doctor_id) DO UPDATE SET
			verified_by = EXCLUDED.verified_by,
			notes = EXCLUDED.notes,
			status = EXCLUDED.status,
			verified_at = EXCLUDED.verified_at,
			updated_at = EXCLUDED.updated_at
		RETURNING doctor_id, verified_by, notes, status, verified_at, created_at, updated_at
	`

	row := r.db.QueryRow(query, doctorID, verifiedBy, notes, now)
	var verification model.DoctorVerification
	if err := row.Scan(&verification.DoctorID, &verification.VerifiedBy, &verification.Notes, &verification.Status, &verification.VerifiedAt, &verification.CreatedAt, &verification.UpdatedAt); err != nil {
		return nil, fmt.Errorf("repository.VerifyDoctor: %w", err)
	}

	if _, err := r.db.Exec(`UPDATE admin_users SET is_verified = TRUE, updated_at = NOW() WHERE id = $1`, doctorID); err != nil {
		return nil, fmt.Errorf("repository.VerifyDoctor update user: %w", err)
	}

	return &verification, nil
}

func (r *AdminRepository) DeactivateUser(userID string) error {
	result, err := r.db.Exec(`UPDATE admin_users SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("repository.DeactivateUser: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.DeactivateUser rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *AdminRepository) SetUserActive(userID string, isActive bool) error {
	result, err := r.db.Exec(`UPDATE admin_users SET is_active = $1, updated_at = NOW() WHERE id = $2`, isActive, userID)
	if err != nil {
		return fmt.Errorf("repository.SetUserActive: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.SetUserActive rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *AdminRepository) SetAuthUserActive(authDatabaseURL, userID string, isActive bool) error {
	if authDatabaseURL == "" {
		return nil
	}

	authDB, err := sql.Open("postgres", authDatabaseURL)
	if err != nil {
		return fmt.Errorf("repository.SetAuthUserActive open auth db: %w", err)
	}
	defer authDB.Close()

	result, err := authDB.Exec(`UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2`, isActive, userID)
	if err != nil {
		return fmt.Errorf("repository.SetAuthUserActive update: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.SetAuthUserActive rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *AdminRepository) SetAuthUserVerified(authDatabaseURL, userID string, isVerified bool) error {
	if authDatabaseURL == "" {
		return nil
	}

	authDB, err := sql.Open("postgres", authDatabaseURL)
	if err != nil {
		return fmt.Errorf("repository.SetAuthUserVerified open auth db: %w", err)
	}
	defer authDB.Close()

	result, err := authDB.Exec(`UPDATE users SET is_verified = $1, updated_at = NOW() WHERE id = $2`, isVerified, userID)
	if err != nil {
		return fmt.Errorf("repository.SetAuthUserVerified update: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.SetAuthUserVerified rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *AdminRepository) ListAppointments() ([]model.Appointment, error) {
	rows, err := r.db.Query(`
		SELECT id, patient_id, doctor_id, status, scheduled_at, reason, created_at, updated_at
		FROM appointments
		ORDER BY scheduled_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("repository.ListAppointments query: %w", err)
	}
	defer rows.Close()

	var appointments []model.Appointment
	for rows.Next() {
		var appointment model.Appointment
		if err := rows.Scan(&appointment.ID, &appointment.PatientID, &appointment.DoctorID, &appointment.Status, &appointment.ScheduledAt, &appointment.Reason, &appointment.CreatedAt, &appointment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository.ListAppointments scan: %w", err)
		}
		appointments = append(appointments, appointment)
	}
	return appointments, rows.Err()
}

func (r *AdminRepository) ListTransactions() ([]model.Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, amount, currency, status, provider, reference, created_at, updated_at
		FROM transactions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("repository.ListTransactions query: %w", err)
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Currency, &transaction.Status, &transaction.Provider, &transaction.Reference, &transaction.CreatedAt, &transaction.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository.ListTransactions scan: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, rows.Err()
}

func (r *AdminRepository) UpsertAppointment(appointment *model.Appointment) error {
	query := `
		INSERT INTO appointments (id, patient_id, doctor_id, status, scheduled_at, reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			patient_id = EXCLUDED.patient_id,
			doctor_id = EXCLUDED.doctor_id,
			status = EXCLUDED.status,
			scheduled_at = EXCLUDED.scheduled_at,
			reason = EXCLUDED.reason,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, appointment.ID, appointment.PatientID, appointment.DoctorID, appointment.Status, appointment.ScheduledAt, appointment.Reason, appointment.CreatedAt, appointment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository.UpsertAppointment: %w", err)
	}
	return nil
}

func (r *AdminRepository) UpsertTransaction(transaction *model.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, amount, currency, status, provider, reference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			status = EXCLUDED.status,
			provider = EXCLUDED.provider,
			reference = EXCLUDED.reference,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, transaction.ID, transaction.UserID, transaction.Amount, transaction.Currency, transaction.Status, transaction.Provider, transaction.Reference, transaction.CreatedAt, transaction.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository.UpsertTransaction: %w", err)
	}
	return nil
}

func (r *AdminRepository) SyncUsersFromAuthDB(authDatabaseURL string) (int, error) {
	if authDatabaseURL == "" {
		return 0, nil
	}

	authDB, err := sql.Open("postgres", authDatabaseURL)
	if err != nil {
		return 0, fmt.Errorf("repository.SyncUsersFromAuthDB open auth db: %w", err)
	}
	defer authDB.Close()

	rows, err := authDB.Query(`
		SELECT id, email, role, first_name, last_name, is_verified, is_active, created_at, updated_at
		FROM users
	`)
	if err != nil {
		return 0, fmt.Errorf("repository.SyncUsersFromAuthDB query users: %w", err)
	}
	defer rows.Close()

	synced := 0
	for rows.Next() {
		var user model.User
		var role string
		if err := rows.Scan(&user.ID, &user.Email, &role, &user.FirstName, &user.LastName, &user.IsVerified, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return synced, fmt.Errorf("repository.SyncUsersFromAuthDB scan: %w", err)
		}
		user.Role = model.Role(role)
		if err := r.UpsertUser(&user); err != nil {
			return synced, fmt.Errorf("repository.SyncUsersFromAuthDB upsert user: %w", err)
		}
		synced++
	}

	if err := rows.Err(); err != nil {
		return synced, fmt.Errorf("repository.SyncUsersFromAuthDB rows: %w", err)
	}

	return synced, nil
}
