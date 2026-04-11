package repository

import (
	"database/sql"
	"fmt"
	"healthcare-platform/services/payment-service/internal/model"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(p *model.Payment) error {
	query := `
		INSERT INTO payments (id, appointment_id, patient_id, amount, currency, status, provider, provider_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(query, p.ID, p.AppointmentID, p.PatientID, p.Amount, p.Currency, p.Status, p.Provider, p.ProviderID, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindByID(id string) (*model.Payment, error) {
	query := `
		SELECT id, appointment_id, patient_id, amount, currency, status, provider, provider_id, created_at, updated_at
		FROM payments
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	var p model.Payment
	err := row.Scan(
		&p.ID, &p.AppointmentID, &p.PatientID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.ProviderID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}
	return &p, nil
}

func (r *PaymentRepository) UpdateStatus(id string, status model.PaymentStatus, providerID string) error {
	query := `
		UPDATE payments
		SET status = $1, provider_id = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, status, providerID, id)
	if err != nil {
		return fmt.Errorf("repository.UpdateStatus: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindByAppointmentID(appointmentID string) (*model.Payment, error) {
	query := `
		SELECT id, appointment_id, patient_id, amount, currency, status, provider, provider_id, created_at, updated_at
		FROM payments
		WHERE appointment_id = $1
	`
	row := r.db.QueryRow(query, appointmentID)
	var p model.Payment
	err := row.Scan(
		&p.ID, &p.AppointmentID, &p.PatientID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.ProviderID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByAppointmentID: %w", err)
	}
	return &p, nil
}
