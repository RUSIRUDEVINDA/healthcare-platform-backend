package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"healthcare-platform/services/patient-service/internal/model"
	"healthcare-platform/services/patient-service/internal/repository"
	"healthcare-platform/pkg/logger"
)

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
	if err := s.repo.Update(userID, req); err != nil {
		return fmt.Errorf("service.UpdateProfile: %w", err)
	}
	s.log.Info("Patient profile updated", "user_id", userID)
	return nil
}
