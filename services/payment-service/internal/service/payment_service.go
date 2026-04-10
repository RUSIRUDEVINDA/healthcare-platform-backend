package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"healthcare-platform/services/payment-service/internal/model"
	"healthcare-platform/services/payment-service/internal/repository"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
)

type PaymentService struct {
	repo     *repository.PaymentRepository
	mqClient *rabbitmq.Client
	log      *logger.Logger
}

type PaymentAlreadyExistsError struct {
	AppointmentID string
	ExistingID    string
}

func (e *PaymentAlreadyExistsError) Error() string {
	return "payment already exists for appointment"
}

func NewPaymentService(repo *repository.PaymentRepository, mqClient *rabbitmq.Client, log *logger.Logger) *PaymentService {
	return &PaymentService{repo: repo, mqClient: mqClient, log: log}
}

func (s *PaymentService) CreatePayment(req *model.CreatePaymentRequest) (*model.PaymentResponse, error) {
	// Idempotency: only one payment per appointment.
	// If a payment already exists for this appointment:
	// - If request matches existing record, return it (safe retry).
	// - If request differs, return 409 Conflict instead of returning "wrong" data.
	existing, err := s.repo.FindByAppointmentID(req.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("service.CreatePayment find existing: %w", err)
	}
	if existing != nil {
		if existing.PatientID != req.PatientID || existing.Amount != req.Amount || existing.Currency != req.Currency {
			return nil, &PaymentAlreadyExistsError{
				AppointmentID: req.AppointmentID,
				ExistingID:    existing.ID,
			}
		}

		return &model.PaymentResponse{PaymentID: existing.ID, Status: string(existing.Status)}, nil
	}

	now := time.Now().UTC()
	p := &model.Payment{
		ID:            uuid.New().String(),
		AppointmentID: req.AppointmentID,
		PatientID:     req.PatientID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        model.StatusPending,
		Provider:      "stripe", // Default for now
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(p); err != nil {
		// Best-effort: if create failed due to a race/duplicate appointment, return existing.
		existing, findErr := s.repo.FindByAppointmentID(req.AppointmentID)
		if findErr == nil && existing != nil {
			if existing.PatientID != req.PatientID || existing.Amount != req.Amount || existing.Currency != req.Currency {
				return nil, &PaymentAlreadyExistsError{
					AppointmentID: req.AppointmentID,
					ExistingID:    existing.ID,
				}
			}
			return &model.PaymentResponse{PaymentID: existing.ID, Status: string(existing.Status)}, nil
		}
		return nil, fmt.Errorf("service.CreatePayment repo: %w", err)
	}

	// Mocking a successful payment for now
	// In reality, this would initiate a call to Stripe's PaymentIntent API
	providerID := "pi_" + uuid.New().String()
	if err := s.repo.UpdateStatus(p.ID, model.StatusCompleted, providerID); err != nil {
		return nil, fmt.Errorf("service.CreatePayment update status: %w", err)
	}

	s.log.Info("Payment created and completed", "payment_id", p.ID, "appointment_id", req.AppointmentID)

	// Publish success event
	s.publishPaymentCompleted(p.ID, req.AppointmentID, providerID)

	return &model.PaymentResponse{
		PaymentID: p.ID,
		Status:    string(model.StatusCompleted),
	}, nil
}

func (s *PaymentService) GetPaymentByID(id string) (*model.Payment, error) {
	return s.repo.FindByID(id)
}

func (s *PaymentService) publishPaymentCompleted(paymentID, appointmentID, providerID string) {
	event := rabbitmq.PaymentCompletedEvent{
		PaymentID:     paymentID,
		AppointmentID: appointmentID,
		ProviderID:    providerID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.mqClient.PublishPaymentCompleted(event); err != nil {
		s.log.Error("Failed to publish payment.completed event", "payment_id", paymentID, "error", err)
	} else {
		s.log.Info("Published payment.completed event", "payment_id", paymentID)
	}
}
