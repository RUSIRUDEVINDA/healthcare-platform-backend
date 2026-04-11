package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/patient-service/internal/model"
	"healthcare-platform/services/patient-service/internal/repository"
)

var ErrPatientNotFound = errors.New("patient not found")

type PatientService struct {
	repo *repository.PatientRepository
	log  *logger.Logger
}

func NewPatientService(repo *repository.PatientRepository, log *logger.Logger) *PatientService {
	return &PatientService{repo: repo, log: log}
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

func (s *PatientService) UpdateProfile(userID string, req *model.UpdatePatientRequest) error {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("service.UpdateProfile find: %w", err)
	}
	if existing == nil {
		return ErrPatientNotFound
	}

	if err := s.repo.Update(userID, req); err != nil {
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

	if err := s.repo.UpdatePartial(userID, req); err != nil {
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

	deleted, err := s.repo.DeleteByUserID(userID)
	if err != nil {
		return fmt.Errorf("service.DeleteProfile: %w", err)
	}
	if !deleted {
		return ErrPatientNotFound
	}

	s.log.Info("Patient profile deleted", "user_id", userID)
	return nil
}
