package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/payment-service/internal/model"
)

type PaymentService struct {
	repo      PaymentRepo
	provider  PaymentProvider
	publisher EventPublisher
	log       *logger.Logger
}

type PaymentRepo interface {
	Create(p *model.Payment) error
	FindByID(id string) (*model.Payment, error)
	UpdateStatus(id string, status model.PaymentStatus, providerID string) error
	FindByAppointmentID(appointmentID string) (*model.Payment, error)
}

type PaymentProvider interface {
	Name() string
	BuildCheckout(p *model.Payment, req *model.CheckoutRequest) (*model.CheckoutResponse, error)
	VerifyNotification(n *model.PayHereNotification) (bool, error)
	MapStatus(statusCode int) model.PaymentStatus
}

type EventPublisher interface {
	PublishPaymentCompleted(event PaymentCompletedEvent) error
}

type PaymentCompletedEvent struct {
	PaymentID     string
	AppointmentID string
	ProviderID    string
	Timestamp     string
}

type PaymentAlreadyExistsError struct {
	AppointmentID string
	ExistingID    string
}

func (e *PaymentAlreadyExistsError) Error() string {
	return "payment already exists for appointment"
}

func NewPaymentService(repo PaymentRepo, provider PaymentProvider, publisher EventPublisher, log *logger.Logger) *PaymentService {
	return &PaymentService{
		repo:      repo,
		provider:  provider,
		publisher: publisher,
		log:       log,
	}
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
		Provider:      s.provider.Name(),
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

	s.log.Info("Payment created (pending)", "payment_id", p.ID, "appointment_id", req.AppointmentID)
	return &model.PaymentResponse{PaymentID: p.ID, Status: string(model.StatusPending), Provider: p.Provider}, nil
}

func (s *PaymentService) GetPaymentByID(id string) (*model.Payment, error) {
	return s.repo.FindByID(id)
}

func (s *PaymentService) Checkout(req *model.CheckoutRequest) (*model.CheckoutResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("service.Checkout: request is nil")
	}

	var p *model.Payment
	var err error
	switch {
	case req.PaymentID != "":
		p, err = s.repo.FindByID(req.PaymentID)
	case req.AppointmentID != "":
		p, err = s.repo.FindByAppointmentID(req.AppointmentID)
	default:
		return nil, fmt.Errorf("service.Checkout: payment_id or appointment_id is required")
	}
	if err != nil {
		return nil, fmt.Errorf("service.Checkout: find payment: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("service.Checkout: payment not found")
	}
	if p.Status == model.StatusCompleted {
		return nil, fmt.Errorf("service.Checkout: payment already completed")
	}

	resp, err := s.provider.BuildCheckout(p, req)
	if err != nil {
		return nil, fmt.Errorf("service.Checkout: provider: %w", err)
	}
	return resp, nil
}

func (s *PaymentService) HandlePayHereNotification(n *model.PayHereNotification) error {
	ok, err := s.provider.VerifyNotification(n)
	if err != nil {
		return fmt.Errorf("service.HandlePayHereNotification verify: %w", err)
	}
	if !ok {
		return fmt.Errorf("service.HandlePayHereNotification: invalid signature")
	}

	p, err := s.repo.FindByID(n.OrderID)
	if err != nil {
		return fmt.Errorf("service.HandlePayHereNotification find payment: %w", err)
	}
	if p == nil {
		return fmt.Errorf("service.HandlePayHereNotification: payment not found")
	}

	expectedAmount := fmt.Sprintf("%.2f", p.Amount)
	if n.PayHereAmount != "" && n.PayHereAmount != expectedAmount {
		return fmt.Errorf("service.HandlePayHereNotification: amount mismatch")
	}
	if n.PayHereCurrency != "" && n.PayHereCurrency != p.Currency {
		return fmt.Errorf("service.HandlePayHereNotification: currency mismatch")
	}

	newStatus := s.provider.MapStatus(n.StatusCode)
	previousStatus := p.Status

	if err := s.repo.UpdateStatus(p.ID, newStatus, n.PaymentID); err != nil {
		return fmt.Errorf("service.HandlePayHereNotification update status: %w", err)
	}

	if previousStatus != model.StatusCompleted && newStatus == model.StatusCompleted {
		s.log.Info("Payment completed via PayHere", "payment_id", p.ID, "appointment_id", p.AppointmentID, "provider_payment_id", n.PaymentID)
		s.publishPaymentCompleted(p.ID, p.AppointmentID, n.PaymentID)
	}

	return nil
}

func (s *PaymentService) publishPaymentCompleted(paymentID, appointmentID, providerID string) {
	event := PaymentCompletedEvent{
		PaymentID:     paymentID,
		AppointmentID: appointmentID,
		ProviderID:    providerID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.publisher.PublishPaymentCompleted(event); err != nil {
		s.log.Error("Failed to publish payment.completed event", "payment_id", paymentID, "error", err)
	} else {
		s.log.Info("Published payment.completed event", "payment_id", paymentID)
	}
}
