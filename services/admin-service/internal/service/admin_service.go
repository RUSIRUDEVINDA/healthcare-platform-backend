package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/admin-service/internal/model"
	"healthcare-platform/services/admin-service/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

type AdminService struct {
	repo            *repository.AdminRepository
	log             *logger.Logger
	authDatabaseURL string
}

func NewAdminService(repo *repository.AdminRepository, log *logger.Logger, authDatabaseURL string) *AdminService {
	return &AdminService{repo: repo, log: log, authDatabaseURL: authDatabaseURL}
}

func (s *AdminService) HandleUserRegistered(event rabbitmq.UserRegisteredEvent) error {
	now := time.Now().UTC()
	user := &model.User{
		ID:         event.UserID,
		Email:      event.Email,
		Role:       model.Role(event.Role),
		FirstName:  event.FirstName,
		LastName:   event.LastName,
		IsVerified: false,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.UpsertUser(user); err != nil {
		return fmt.Errorf("service.HandleUserRegistered: %w", err)
	}

	s.log.Info("Admin user mirror updated from user.registered", "user_id", event.UserID, "role", event.Role)
	return nil
}

func (s *AdminService) HandleAppointmentBooked(event rabbitmq.AppointmentBookedEvent) error {
	scheduledAt, err := time.Parse(time.RFC3339, event.ScheduledAt)
	if err != nil {
		scheduledAt = time.Now().UTC()
	}

	appointment := &model.Appointment{
		ID:          event.AppointmentID,
		PatientID:   event.PatientID,
		DoctorID:    event.DoctorID,
		Status:      "booked",
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.UpsertAppointment(appointment); err != nil {
		return fmt.Errorf("service.HandleAppointmentBooked: %w", err)
	}

	s.log.Info("Admin appointment mirror updated", "appointment_id", event.AppointmentID)
	return nil
}

func (s *AdminService) HandlePaymentCompleted(event rabbitmq.PaymentCompletedEvent) error {
	transaction := &model.Transaction{
		ID:        event.TransactionID,
		UserID:    event.UserID,
		Amount:    event.Amount,
		Currency:  event.Currency,
		Status:    event.Status,
		Provider:  event.Provider,
		Reference: event.Reference,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.UpsertTransaction(transaction); err != nil {
		return fmt.Errorf("service.HandlePaymentCompleted: %w", err)
	}

	s.log.Info("Admin transaction mirror updated", "transaction_id", event.TransactionID)
	return nil
}

func (s *AdminService) ListUsers() ([]model.User, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *AdminService) ListAppointments() ([]model.Appointment, error) {
	appointments, err := s.repo.ListAppointments()
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func (s *AdminService) ListTransactions() ([]model.Transaction, error) {
	transactions, err := s.repo.ListTransactions()
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *AdminService) VerifyDoctor(doctorID, verifiedBy, notes string) (*model.DoctorVerification, error) {
	verification, err := s.repo.VerifyDoctor(doctorID, verifiedBy, notes)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

func (s *AdminService) DeactivateUser(userID string) error {
	if userID == "" {
		return ErrUserNotFound
	}
	if err := s.repo.DeactivateUser(userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
