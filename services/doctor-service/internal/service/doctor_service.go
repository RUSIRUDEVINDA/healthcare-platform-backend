package service

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/doctor-service/internal/model"
	"healthcare-platform/services/doctor-service/internal/repository"
)

var (
	ErrDoctorNotFound     = errors.New("doctor not found")
	ErrDuplicateNICOrSLMC = errors.New("NIC or SLMC number already in use")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
	ErrInvalidNICFormat   = errors.New("nic must be exactly 12 digits")
	ErrInvalidSLMCFormat  = errors.New("slmc_no must be exactly 5 digits")
)

// IdentityInUseError is returned on update when NIC or SLMC matches another doctor's record.
type IdentityInUseError struct {
	Field         string
	OtherDoctorID int64
	OmitJSONKey   string
}

func (e *IdentityInUseError) Error() string {
	return fmt.Sprintf(
		`this %s is already registered to doctor id %d; omit "%s" from the JSON body if you are not changing it`,
		e.Field, e.OtherDoctorID, e.OmitJSONKey,
	)
}

var nicDigitsRE = regexp.MustCompile(`^[0-9]{12}$`)
var slmcDigitsRE = regexp.MustCompile(`^[0-9]{5}$`)

// DoctorService contains business logic for doctor management.
type DoctorService struct {
	repo *repository.DoctorRepository
	mq   *rabbitmq.Client
	log  *logger.Logger
}

func NewDoctorService(repo *repository.DoctorRepository, mq *rabbitmq.Client, log *logger.Logger) *DoctorService {
	return &DoctorService{repo: repo, mq: mq, log: log}
}

func (s *DoctorService) List(specialization string) ([]model.Doctor, error) {
	return s.repo.List(specialization)
}

func (s *DoctorService) GetByID(id int64) (*model.Doctor, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, ErrDoctorNotFound
	}
	return d, nil
}

func (s *DoctorService) Create(req *model.CreateDoctorRequest) (*model.Doctor, error) {
	nic := strings.TrimSpace(req.NIC)
	slmc := strings.TrimSpace(req.SLMCNo)
	if !nicDigitsRE.MatchString(nic) {
		return nil, ErrInvalidNICFormat
	}
	if !slmcDigitsRE.MatchString(slmc) {
		return nil, ErrInvalidSLMCFormat
	}

	d := &model.Doctor{
		Name:           req.Name,
		Specialization: req.Specialization,
		Experience:     req.Experience,
		Hospital:       req.Hospital,
		NIC:            nic,
		SLMCNo:         slmc,
	}
	if err := s.repo.Create(d); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateNICOrSLMC
		}
		return nil, fmt.Errorf("service.Create: %w", err)
	}

	s.log.Info("Doctor created", "doctor_id", d.ID, "specialization", d.Specialization)

	if s.mq != nil {
		ev := rabbitmq.DoctorCreatedEvent{
			DoctorID:       uint(d.ID),
			Name:           d.Name,
			Specialization: d.Specialization,
			Hospital:       d.Hospital,
			NIC:            d.NIC,
			SLMCNo:         d.SLMCNo,
		}
		if err := s.mq.PublishDoctorCreated(ev); err != nil {
			s.log.Warn("Failed to publish doctor.created event", "error", err)
		}
	}

	return d, nil
}

func (s *DoctorService) Update(id int64, req *model.UpdateDoctorRequest) (*model.Doctor, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrDoctorNotFound
	}

	if req.Name == nil && req.Specialization == nil && req.Experience == nil &&
		req.Hospital == nil && req.NIC == nil && req.SLMCNo == nil {
		return nil, ErrNoFieldsToUpdate
	}

	if req.NIC != nil {
		trimmed := strings.TrimSpace(*req.NIC)
		req.NIC = &trimmed
	}
	if req.SLMCNo != nil {
		trimmed := strings.TrimSpace(*req.SLMCNo)
		req.SLMCNo = &trimmed
	}

	if req.NIC != nil && !nicDigitsRE.MatchString(*req.NIC) {
		return nil, ErrInvalidNICFormat
	}
	if req.SLMCNo != nil && !slmcDigitsRE.MatchString(*req.SLMCNo) {
		return nil, ErrInvalidSLMCFormat
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Specialization != nil {
		existing.Specialization = *req.Specialization
	}
	if req.Experience != nil {
		existing.Experience = *req.Experience
	}
	if req.Hospital != nil {
		existing.Hospital = *req.Hospital
	}
	if req.NIC != nil {
		existing.NIC = *req.NIC
	}
	if req.SLMCNo != nil {
		existing.SLMCNo = *req.SLMCNo
	}

	if otherID, taken, err := s.repo.OtherDoctorIDWithNIC(id, existing.NIC); err != nil {
		return nil, err
	} else if taken {
		return nil, &IdentityInUseError{Field: "NIC", OtherDoctorID: otherID, OmitJSONKey: "nic"}
	}
	if otherID, taken, err := s.repo.OtherDoctorIDWithSLMC(id, existing.SLMCNo); err != nil {
		return nil, err
	} else if taken {
		return nil, &IdentityInUseError{Field: "SLMC number", OtherDoctorID: otherID, OmitJSONKey: "slmc_no"}
	}

	if err := s.repo.Update(existing); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDoctorNotFound
		}
		if isUniqueViolation(err) {
			return nil, ErrDuplicateNICOrSLMC
		}
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	d, err := s.repo.FindByID(id)
	if err != nil || d == nil {
		return nil, ErrDoctorNotFound
	}
	return d, nil
}

func (s *DoctorService) Delete(id int64) error {
	n, err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("service.Delete: %w", err)
	}
	if n == 0 {
		return ErrDoctorNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
