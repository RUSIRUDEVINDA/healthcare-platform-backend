package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/admin-service/internal/model"
	"healthcare-platform/services/admin-service/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

type AdminService struct {
	repo            *repository.AdminRepository
	log                   *logger.Logger
	authDatabaseURL       string
	appointmentDatabaseURL string
	paymentDatabaseURL     string
}

func NewAdminService(repo *repository.AdminRepository, log *logger.Logger, authDB, appDB, payDB string) *AdminService {
	return &AdminService{repo: repo, log: log, authDatabaseURL: authDB, appointmentDatabaseURL: appDB, paymentDatabaseURL: payDB}
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
	if s.paymentDatabaseURL == "" {
		s.log.Info("Admin payment completed event received but payment DB URL is empty",
			"payment_id", event.PaymentID,
			"appointment_id", event.AppointmentID,
			"provider_id", event.ProviderID,
		)
		return nil
	}

	synced, err := s.repo.SyncPaymentFromDB(s.paymentDatabaseURL, event.PaymentID, event.AppointmentID)
	if err != nil {
		return fmt.Errorf("service.HandlePaymentCompleted sync: %w", err)
	}
	if !synced {
		s.log.Warn("Payment completed event received but matching payment row was not found",
			"payment_id", event.PaymentID,
			"appointment_id", event.AppointmentID,
			"provider_id", event.ProviderID,
		)
		return nil
	}

	s.log.Info("Admin payment completed event recorded",
		"payment_id", event.PaymentID,
		"appointment_id", event.AppointmentID,
		"provider_id", event.ProviderID,
	)
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

func (s *AdminService) ReactivateUser(userID string) error {
	if userID == "" {
		return ErrUserNotFound
	}
	if err := s.repo.ReactivateUser(userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (s *AdminService) SyncAll() (int, error) {
	s.log.Info("Starting manual sync from other services...")
	
	usersSynced, err := s.repo.SyncUsersFromAuthDB(s.authDatabaseURL)
	if err != nil {
		s.log.Error("User sync failed", "error", err)
	}

	appsSynced, err := s.repo.SyncAppointmentsFromDB(s.appointmentDatabaseURL)
	if err != nil {
		s.log.Error("Appointment sync failed", "error", err)
	}

	txSynced, err := s.repo.SyncTransactionsFromDB(s.paymentDatabaseURL)
	if err != nil {
		s.log.Error("Transaction sync failed", "error", err)
	}

	total := usersSynced + appsSynced + txSynced
	s.log.Info("Sync completed", "users", usersSynced, "appointments", appsSynced, "transactions", txSynced)
	return total, nil
}

func (s *AdminService) CreateUser(req model.CreateUserRequest) (*model.User, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	userID := uuid.New().String()
	now := time.Now().UTC()
	user := &model.User{
		ID:         userID,
		Email:      req.Email,
		Role:       req.Role,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		IsVerified: req.Role == model.RoleAdmin, // Admins auto-verified
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 1. Authoritative Create
	if err := s.repo.CreateUserInAuthDB(s.authDatabaseURL, user, string(hashedPwd)); err != nil {
		return nil, fmt.Errorf("service.CreateUser auth: %w", err)
	}

	// 2. Mirror Update
	if err := s.repo.UpsertUser(user); err != nil {
		s.log.Error("CreateUser mirror update failed", "error", err)
	}

	return user, nil
}

func (s *AdminService) UpdateUser(userID string, req model.UpdateUserRequest) error {
	// 1. Authoritative Update
	if err := s.repo.UpdateUserInAuthDB(s.authDatabaseURL, userID, req.FirstName, req.LastName, req.Email); err != nil {
		return fmt.Errorf("service.UpdateUser auth: %w", err)
	}

	// 2. Mirror Update (Sync Single)
	// We only have partial info here for Upsert, sync from DB is better to keep all fields consistent
	if _, err := s.repo.SyncUsersFromAuthDB(s.authDatabaseURL); err != nil {
		s.log.Error("UpdateUser sync failed", "error", err)
	}

	return nil
}

func (s *AdminService) CreateAppointment(req model.CreateAppointmentRequest) (*model.Appointment, error) {
	apptID := uuid.New().String()
	now := time.Now().UTC()
	appt := &model.Appointment{
		ID:          apptID,
		PatientID:   req.PatientID,
		DoctorID:    req.DoctorID,
		Status:      "confirmed",
		ScheduledAt: req.ScheduledAt,
		Reason:      req.Reason,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 1. Authoritative Create
	if err := s.repo.CreateAppointmentInSourceDB(s.appointmentDatabaseURL, appt); err != nil {
		return nil, fmt.Errorf("service.CreateAppointment appt: %w", err)
	}

	// 2. Mirror Update
	if err := s.repo.UpsertAppointment(appt); err != nil {
		s.log.Error("CreateAppointment mirror update failed", "error", err)
	}

	return appt, nil
}

func (s *AdminService) UpdateAppointment(apptID string, req model.UpdateAppointmentRequest) error {
	// 1. Authoritative Update
	if err := s.repo.UpdateAppointmentInSourceDB(s.appointmentDatabaseURL, apptID, req.Status, req.ScheduledAt, req.Reason); err != nil {
		return fmt.Errorf("service.UpdateAppointment appt: %w", err)
	}

	// 2. Mirror Update (Sync Single)
	if _, err := s.repo.SyncAppointmentsFromDB(s.appointmentDatabaseURL); err != nil {
		s.log.Error("UpdateAppointment sync failed", "error", err)
	}

	return nil
}

func (s *AdminService) CancelAppointment(apptID string) error {
	// 1. Authoritative Cancel
	if err := s.repo.CancelAppointmentInSourceDB(s.appointmentDatabaseURL, apptID); err != nil {
		return fmt.Errorf("service.CancelAppointment appt: %w", err)
	}

	// 2. Mirror Update
	if _, err := s.repo.SyncAppointmentsFromDB(s.appointmentDatabaseURL); err != nil {
		s.log.Error("CancelAppointment sync failed", "error", err)
	}

	return nil
}
