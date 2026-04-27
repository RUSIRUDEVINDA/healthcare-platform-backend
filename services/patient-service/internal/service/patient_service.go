package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/patient-service/internal/model"
	"healthcare-platform/services/patient-service/internal/repository"
)

var ErrPatientNotFound = errors.New("patient not found")
var ErrPatientNICAlreadyExists = errors.New("NIC already exists")

type PatientService struct {
	repo *repository.PatientRepository
	mq   *rabbitmq.Client
	log  *logger.Logger
}

func NewPatientService(repo *repository.PatientRepository, mq *rabbitmq.Client, log *logger.Logger) *PatientService {
	return &PatientService{repo: repo, mq: mq, log: log}
}

func (s *PatientService) CreateFromUserEvent(userID, email, firstName, lastName string) error {
	now := time.Now().UTC()
	p := &model.Patient{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(p); err != nil {
		return fmt.Errorf("service.CreateFromUserEvent: %w", err)
	}

	s.log.Info("Patient profile created from event", "user_id", userID, "email", email)
	return nil
}

func (s *PatientService) GetProfile(userID string) (*model.Patient, error) {
	p, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetProfile: %w", err)
	}
	return p, nil
}

// GetDisplayByUserID returns stored patient profile names for a user (for internal service callers).
func (s *PatientService) GetDisplayByUserID(userID string) (firstName, lastName string, found bool, err error) {
	p, e := s.repo.FindByUserID(userID)
	if e != nil {
		return "", "", false, fmt.Errorf("service.GetDisplayByUserID: %w", e)
	}
	if p == nil {
		return "", "", false, nil
	}
	return p.FirstName, p.LastName, true, nil
}

func (s *PatientService) EnsureProfile(userID, email, firstName, lastName string) (*model.Patient, error) {
	p, err := s.GetProfile(userID)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}

	s.log.Info("Patient profile not found during GetProfile, attempting lazy creation", "user_id", userID)
	if err := s.CreateFromUserEvent(userID, email, firstName, lastName); err != nil {
		return nil, fmt.Errorf("service.EnsureProfile: %w", err)
	}
	return s.GetProfile(userID)
}

func (s *PatientService) UpdateProfile(userID string, req *model.UpdatePatientRequest) error {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("service.UpdateProfile find: %w", err)
	}
	if existing == nil {
		return ErrPatientNotFound
	}

	normalizeOptionalString(&req.NIC)

	if err := s.repo.Update(userID, req); err != nil {
		if isUniqueViolation(err) {
			return ErrPatientNICAlreadyExists
		}
		return fmt.Errorf("service.UpdateProfile: %w", err)
	}
	s.log.Info("Patient profile updated", "user_id", userID)
	return nil
}

func (s *PatientService) PatchProfile(userID string, req *model.PatchPatientRequest) error {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("service.PatchProfile find: %w", err)
	}
	if existing == nil {
		return ErrPatientNotFound
	}

	normalizeOptionalString(&req.NIC)

	if err := s.repo.UpdatePartial(userID, req); err != nil {
		if isUniqueViolation(err) {
			return ErrPatientNICAlreadyExists
		}
		return fmt.Errorf("service.PatchProfile: %w", err)
	}
	s.log.Info("Patient profile patched", "user_id", userID)
	return nil
}

func (s *PatientService) DeleteProfile(userID string) error {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("service.DeleteProfile find: %w", err)
	}
	if existing == nil {
		return ErrPatientNotFound
	}

	patientID, err := s.repo.DeleteByUserID(userID)
	if err != nil {
		return fmt.Errorf("service.DeleteProfile: %w", err)
	}
	if patientID == "" {
		return ErrPatientNotFound
	}

	s.log.Info("Patient profile deleted", "user_id", userID, "patient_id", patientID)

	// Publish patient.deleted event to notify other services
	event := rabbitmq.PatientDeletedEvent{
		PatientID: patientID,
		UserID:    userID,
	}
	if err := s.mq.PublishPatientDeleted(event); err != nil {
		// Log the error but don't fail the deletion
		s.log.Error("Failed to publish patient.deleted event", "user_id", userID, "patient_id", patientID, "error", err)
	}

	return nil
}

func normalizeOptionalString(value **string) {
	if value == nil || *value == nil {
		return
	}

	trimmed := strings.TrimSpace(**value)
	if trimmed == "" {
		*value = nil
		return
	}

	**value = trimmed
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
