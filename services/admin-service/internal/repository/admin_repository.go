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

func (r *AdminRepository) ReactivateUser(userID string) error {
	result, err := r.db.Exec(`UPDATE admin_users SET is_active = TRUE, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("repository.ReactivateUser: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.ReactivateUser rows: %w", err)
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

// Authoritative CRUD operations on source databases

func (r *AdminRepository) CreateUserInAuthDB(authDatabaseURL string, user *model.User, passwordHash string) error {
	db, err := sql.Open("postgres", authDatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO users (id, email, password_hash, role, first_name, last_name, is_verified, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = db.Exec(query, user.ID, user.Email, passwordHash, string(user.Role), user.FirstName, user.LastName, user.IsVerified, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *AdminRepository) UpdateUserInAuthDB(authDatabaseURL string, userID string, firstName, lastName, email string) error {
	db, err := sql.Open("postgres", authDatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `UPDATE users SET first_name = $1, last_name = $2, email = $3, updated_at = NOW() WHERE id = $4`
	_, err = db.Exec(query, firstName, lastName, email, userID)
	return err
}

func (r *AdminRepository) CreateAppointmentInSourceDB(appointmentDatabaseURL string, appt *model.Appointment) error {
	db, err := sql.Open("postgres", appointmentDatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO appointments (id, patient_id, doctor_id, status, scheduled_at, reason, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = db.Exec(query, appt.ID, appt.PatientID, appt.DoctorID, appt.Status, appt.ScheduledAt, appt.Reason, appt.CreatedAt, appt.UpdatedAt)
	return err
}

func (r *AdminRepository) UpdateAppointmentInSourceDB(appointmentDatabaseURL string, apptID string, status string, scheduledAt time.Time, reason string) error {
	db, err := sql.Open("postgres", appointmentDatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `UPDATE appointments SET status = $1, scheduled_at = $2, reason = $3, updated_at = NOW() WHERE id = $4`
	_, err = db.Exec(query, status, scheduledAt, reason, apptID)
	return err
}

func (r *AdminRepository) CancelAppointmentInSourceDB(appointmentDatabaseURL string, apptID string) error {
	db, err := sql.Open("postgres", appointmentDatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `UPDATE appointments SET status = 'cancelled', updated_at = NOW() WHERE id = $1`
	_, err = db.Exec(query, apptID)
	return err
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

func (r *AdminRepository) SyncAppointmentsFromDB(appointmentDatabaseURL string) (int, error) {
	if appointmentDatabaseURL == "" {
		return 0, nil
	}
	db, err := sql.Open("postgres", appointmentDatabaseURL)
	if err != nil {
		return 0, fmt.Errorf("repository.SyncAppointmentsFromDB open db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, patient_id, doctor_id, status, scheduled_at, reason, created_at, updated_at FROM appointments`)
	if err != nil {
		return 0, fmt.Errorf("repository.SyncAppointmentsFromDB query: %w", err)
	}
	defer rows.Close()

	synced := 0
	for rows.Next() {
		var a model.Appointment
		if err := rows.Scan(&a.ID, &a.PatientID, &a.DoctorID, &a.Status, &a.ScheduledAt, &a.Reason, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return synced, err
		}
		if err := r.UpsertAppointment(&a); err != nil {
			return synced, err
		}
		synced++
	}
	return synced, nil
}

func (r *AdminRepository) SyncTransactionsFromDB(paymentDatabaseURL string) (int, error) {
	if paymentDatabaseURL == "" {
		return 0, nil
	}
	db, err := sql.Open("postgres", paymentDatabaseURL)
	if err != nil {
		return 0, fmt.Errorf("repository.SyncTransactionsFromDB open db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, appointment_id, patient_id, amount, currency, status, provider, provider_id, created_at, updated_at
		FROM payments
	`)
	if err != nil {
		return 0, fmt.Errorf("repository.SyncTransactionsFromDB query: %w", err)
	}
	defer rows.Close()

	synced := 0
	for rows.Next() {
		var (
			paymentID     string
			appointmentID string
			patientID     string
			amount        float64
			currency      string
			status        string
			provider      string
			providerID    sql.NullString
			createdAt     time.Time
			updatedAt     time.Time
		)
		if err := rows.Scan(&paymentID, &appointmentID, &patientID, &amount, &currency, &status, &provider, &providerID, &createdAt, &updatedAt); err != nil {
			return synced, err
		}
		t := model.Transaction{
			ID:        paymentID,
			UserID:    patientID,
			Amount:    amount,
			Currency:  currency,
			Status:    status,
			Provider:  provider,
			Reference: appointmentID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		if providerID.Valid && providerID.String != "" {
			t.Reference = providerID.String
		}
		if err := r.UpsertTransaction(&t); err != nil {
			return synced, err
		}
		synced++
	}
	return synced, nil
}

func (r *AdminRepository) SyncPaymentFromDB(paymentDatabaseURL, paymentID, appointmentID string) (bool, error) {
	if paymentDatabaseURL == "" {
		return false, nil
	}

	db, err := sql.Open("postgres", paymentDatabaseURL)
	if err != nil {
		return false, fmt.Errorf("repository.SyncPaymentFromDB open db: %w", err)
	}
	defer db.Close()

	var (
		id            string
		foundApptID    string
		patientID     string
		amount        float64
		currency      string
		status        string
		provider      string
		providerID    sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
	)

	err = db.QueryRow(`
		SELECT id, appointment_id, patient_id, amount, currency, status, provider, provider_id, created_at, updated_at
		FROM payments
		WHERE id = $1 OR appointment_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, paymentID, appointmentID).Scan(&id, &foundApptID, &patientID, &amount, &currency, &status, &provider, &providerID, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("repository.SyncPaymentFromDB query: %w", err)
	}

	transaction := &model.Transaction{
		ID:        id,
		UserID:    patientID,
		Amount:    amount,
		Currency:  currency,
		Status:    status,
		Provider:  provider,
		Reference: foundApptID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	if providerID.Valid && providerID.String != "" {
		transaction.Reference = providerID.String
	}

	if err := r.UpsertTransaction(transaction); err != nil {
		return false, fmt.Errorf("repository.SyncPaymentFromDB upsert: %w", err)
	}
	return true, nil
}
