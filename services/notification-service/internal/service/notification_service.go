package service

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/notification-service/internal/config"
	"healthcare-platform/services/notification-service/internal/model"
)

type NotificationService struct {
	cfg *config.Config
	log *logger.Logger
}

func NewNotificationService(cfg *config.Config, log *logger.Logger) *NotificationService {
	return &NotificationService{cfg: cfg, log: log}
}

func (s *NotificationService) HandleAppointmentBooked(event rabbitmq.AppointmentBookedEvent) error {
	subject := "Appointment Confirmation"
	body := fmt.Sprintf(
		"Your appointment %s is confirmed for %s. Consultation fee: %.2f.",
		event.AppointmentID,
		event.ScheduledAt,
		event.ConsultFee,
	)

	if err := s.deliverEmail(event.PatientEmail, subject, body); err != nil {
		return err
	}
	if event.PatientPhone != "" {
		if err := s.deliverSMS(event.PatientPhone, body); err != nil {
			return err
		}
	}

	s.log.Info("Handled appointment.booked notification", "appointment_id", event.AppointmentID)
	return nil
}

func (s *NotificationService) HandleAppointmentCancelled(event rabbitmq.AppointmentCancelledEvent) error {
	body := fmt.Sprintf("Your appointment %s has been cancelled.", event.AppointmentID)
	if err := s.deliverEmail(event.PatientEmail, "Appointment Cancelled", body); err != nil {
		return err
	}
	if event.PatientPhone != "" {
		if err := s.deliverSMS(event.PatientPhone, body); err != nil {
			return err
		}
	}
	s.log.Info("Handled appointment.cancelled notification", "appointment_id", event.AppointmentID)
	return nil
}

func (s *NotificationService) HandleConsultationCompleted(event rabbitmq.ConsultationCompletedEvent) error {
	body := fmt.Sprintf("Your consultation %s has been marked complete.", event.ConsultationID)
	if err := s.deliverEmail(event.PatientEmail, "Consultation Completed", body); err != nil {
		return err
	}
	if event.PatientPhone != "" {
		if err := s.deliverSMS(event.PatientPhone, body); err != nil {
			return err
		}
	}
	s.log.Info("Handled consultation.completed notification", "consultation_id", event.ConsultationID)
	return nil
}

func (s *NotificationService) deliverEmail(recipient, subject, body string) error {
	payload := model.NotificationPayload{
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
		Channel:   model.DeliveryChannelEmail,
	}

	if s.cfg.SMTPHost == "" || s.cfg.FromEmail == "" {
		s.log.Info("Email provider not configured, logging notification instead", "recipient", recipient, "subject", subject)
		return s.logDelivery(payload, "log")
	}

	message := []byte("From: " + s.cfg.FromEmail + "\r\n" +
		"To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	if err := smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{recipient}, message); err != nil {
		return fmt.Errorf("notification.deliverEmail: %w", err)
	}

	s.log.Info("Email delivered", "recipient", recipient, "subject", subject)
	return s.logDelivery(payload, "smtp")
}

func (s *NotificationService) deliverSMS(recipient, body string) error {
	payload := model.NotificationPayload{
		Recipient: recipient,
		Body:      body,
		Channel:   model.DeliveryChannelSMS,
	}

	if s.cfg.TwilioAccountSID == "" || s.cfg.TwilioAuthToken == "" || s.cfg.TwilioFromNumber == "" {
		s.log.Info("SMS provider not configured, logging notification instead", "recipient", recipient)
		return s.logDelivery(payload, "log")
	}

	form := url.Values{}
	form.Set("To", recipient)
	form.Set("From", s.cfg.TwilioFromNumber)
	form.Set("Body", body)

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.cfg.TwilioAccountSID)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("notification.deliverSMS request: %w", err)
	}
	req.SetBasicAuth(s.cfg.TwilioAccountSID, s.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("notification.deliverSMS execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification.deliverSMS unexpected status: %s", resp.Status)
	}

	s.log.Info("SMS delivered", "recipient", recipient)
	return s.logDelivery(payload, "twilio")
}

func (s *NotificationService) logDelivery(payload model.NotificationPayload, provider string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(payload.Body))
	s.log.Info("Notification recorded", "channel", string(payload.Channel), "provider", provider, "recipient", payload.Recipient, "payload_b64", encoded)
	return nil
}
