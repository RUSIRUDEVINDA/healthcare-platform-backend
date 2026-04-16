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
	if _, err := uuid.Parse(event.PaymentID); err != nil {
		s.log.Warn("Skipping payment.completed mirror due to invalid payment_id", "payment_id", event.PaymentID, "error", err)
		return nil
	}

	transaction := &model.Transaction{
		ID:        event.PaymentID,
		UserID:    uuid.Nil.String(),
		Amount:    0,
		Currency:  "LKR",
		Status:    "completed",
		Provider:  "payment-service",
		Reference: event.ProviderID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.UpsertTransaction(transaction); err != nil {
		return fmt.Errorf("service.HandlePaymentCompleted: %w", err)
	}

	s.log.Info("Admin transaction mirror updated", "transaction_id", event.PaymentID)
	return nil
}

func (s *AdminService) ListUsers() ([]model.User, error) {
	if s.authDatabaseURL != "" {
		synced, err := s.repo.SyncUsersFromAuthDB(s.authDatabaseURL)
		if err != nil {
			s.log.Warn("Failed to sync users from auth DB", "error", err)
		} else if synced > 0 {
			s.log.Info("Synced users from auth DB", "count", synced)
		}
	}

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
	if doctorID == "" {
		return nil, ErrUserNotFound
	}

	if err := s.repo.SetAuthUserVerified(s.authDatabaseURL, doctorID, true); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	verification, err := s.repo.VerifyDoctor(doctorID, verifiedBy, notes)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

func (s *AdminService) DeactivateUser(userID string) error {
	return s.setUserActive(userID, false)
}

func (s *AdminService) ReactivateUser(userID string) error {
	return s.setUserActive(userID, true)
}

func (s *AdminService) setUserActive(userID string, isActive bool) error {
	if userID == "" {
		return ErrUserNotFound
	}

	if err := s.repo.SetAuthUserActive(s.authDatabaseURL, userID, isActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.repo.SetUserActive(userID, isActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if s.authDatabaseURL != "" {
				if _, syncErr := s.repo.SyncUsersFromAuthDB(s.authDatabaseURL); syncErr == nil {
					return nil
				}
			}
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
