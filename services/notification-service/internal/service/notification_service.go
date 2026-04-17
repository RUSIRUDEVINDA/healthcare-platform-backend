package service

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
	"time"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/notification-service/internal/config"
	"healthcare-platform/services/notification-service/internal/model"
)

type NotificationService struct {
	cfg *config.Config
	log *logger.Logger

	mu              sync.Mutex
	appointments    map[string]rabbitmq.AppointmentBookedEvent
	pendingPayments map[string]rabbitmq.PaymentCompletedEvent
}

func NewNotificationService(cfg *config.Config, log *logger.Logger) *NotificationService {
	return &NotificationService{
		cfg:             cfg,
		log:             log,
		appointments:    make(map[string]rabbitmq.AppointmentBookedEvent),
		pendingPayments: make(map[string]rabbitmq.PaymentCompletedEvent),
	}
}

func (s *NotificationService) HandleAppointmentBooked(event rabbitmq.AppointmentBookedEvent) error {
	s.rememberAppointment(event)

	if pending, ok := s.consumePendingPayment(event.AppointmentID); ok {
		return s.sendPaymentConfirmation(event, pending)
	}

	if strings.EqualFold(strings.TrimSpace(event.PaymentStatus), "paid") {
		s.log.Info("Skipping appointment.booked email because payment is already completed", "appointment_id", event.AppointmentID)
		return nil
	}

	subject := "Appointment Confirmation"
	scheduledAt := event.ScheduledAt
	if scheduledAt == "" {
		scheduledAt = event.Time
	}

	htmlBody := s.renderAppointmentConfirmationEmail(event, rabbitmq.PaymentCompletedEvent{}, scheduledAt)
	textBody := fmt.Sprintf(
		"Your appointment %s is confirmed for %s. Consultation fee: %.2f.",
		event.AppointmentID,
		scheduledAt,
		event.ConsultFee,
	)

	if err := s.deliverHTMLEmail(event.PatientEmail, subject, htmlBody, textBody); err != nil {
		return err
	}
	if event.PatientPhone != "" {
		if err := s.deliverSMS(event.PatientPhone, textBody); err != nil {
			return err
		}
	}

	s.log.Info("Handled appointment.booked notification (unpaid)", "appointment_id", event.AppointmentID)
	return nil
}

func (s *NotificationService) HandlePaymentCompleted(event rabbitmq.PaymentCompletedEvent) error {
	if event.AppointmentID == "" {
		return nil
	}

	appt, ok := s.getAppointment(event.AppointmentID)
	if !ok {
		s.storePendingPayment(event)
		s.log.Info("Queued payment confirmation until appointment details are available", "appointment_id", event.AppointmentID, "payment_id", event.PaymentID)
		return nil
	}

	return s.sendPaymentConfirmation(appt, event)
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

func (s *NotificationService) sendPaymentConfirmation(appt rabbitmq.AppointmentBookedEvent, payment rabbitmq.PaymentCompletedEvent) error {
	subject := "Appointment Confirmed"
	scheduledAt := appt.ScheduledAt
	if scheduledAt == "" {
		scheduledAt = appt.Time
	}

	htmlBody := s.renderAppointmentConfirmationEmail(appt, payment, scheduledAt)
	textBody := fmt.Sprintf(
		"Your appointment %s has been confirmed. Scheduled for %s. Consultation fee: %.2f.",
		appt.AppointmentID,
		scheduledAt,
		appt.ConsultFee,
	)

	if err := s.deliverHTMLEmail(appt.PatientEmail, subject, htmlBody, textBody); err != nil {
		return err
	}

	s.log.Info("Handled payment.completed appointment confirmation", "appointment_id", appt.AppointmentID, "payment_id", payment.PaymentID)
	return nil
}

func (s *NotificationService) deliverEmail(recipient, subject, body string) error {
	return s.deliverEmailWithContentType(recipient, subject, body, "text/plain; charset=UTF-8")
}

func (s *NotificationService) deliverHTMLEmail(recipient, subject, htmlBody, textFallback string) error {
	return s.deliverEmailWithContentType(recipient, subject, htmlBody, "text/html; charset=UTF-8", textFallback)
}

func (s *NotificationService) deliverEmailWithContentType(recipient, subject, body, contentType string, plainFallback ...string) error {
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

	if strings.TrimSpace(recipient) == "" {
		s.log.Warn("Skipping email delivery because recipient is empty", "subject", subject)
		return s.logDelivery(payload, "log")
	}

	fromAddr := envelopeAddress(s.cfg.FromEmail)
	headerFrom := s.cfg.FromEmail
	if headerFrom == "" {
		headerFrom = fromAddr
	}

	message := []byte("From: " + headerFrom + "\r\n" +
		"To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: " + contentType + "\r\n\r\n" +
		body + "\r\n")
	if len(plainFallback) > 0 && strings.HasPrefix(contentType, "text/html") {
		message = []byte("From: " + headerFrom + "\r\n" +
			"To: " + recipient + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: multipart/alternative; boundary=notification-boundary\r\n\r\n" +
			"--notification-boundary\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
			plainFallback[0] + "\r\n\r\n" +
			"--notification-boundary\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
			body + "\r\n\r\n" +
			"--notification-boundary--\r\n")
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	if s.cfg.SMTPSecure {
		tlsCfg := &tls.Config{
			ServerName:         s.cfg.SMTPHost,
			InsecureSkipVerify: s.cfg.SMTPSkipVerify,
		}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("notification.deliverEmail tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("notification.deliverEmail smtp client: %w", err)
		}
		defer client.Close()

		if err := sendSMTPMessage(client, auth, fromAddr, recipient, message, false, s.cfg.SMTPHost, s.cfg.SMTPSkipVerify); err != nil {
			return fmt.Errorf("notification.deliverEmail: %w", err)
		}
	} else {
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("notification.deliverEmail dial: %w", err)
		}
		defer client.Close()

		if err := sendSMTPMessage(client, auth, fromAddr, recipient, message, true, s.cfg.SMTPHost, s.cfg.SMTPSkipVerify); err != nil {
			return fmt.Errorf("notification.deliverEmail: %w", err)
		}
	}

	s.log.Info("Email delivered", "recipient", recipient, "subject", subject)
	return s.logDelivery(payload, "smtp")
}

func sendSMTPMessage(client *smtp.Client, auth smtp.Auth, fromAddr, recipient string, message []byte, allowStartTLS bool, serverName string, skipVerify bool) error {
	if allowStartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			// For plain connections, upgrade when the server supports it.
			if err := client.StartTLS(&tls.Config{ServerName: serverName, InsecureSkipVerify: skipVerify}); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(fromAddr); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write(message); err != nil {
		wc.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return nil
}

func envelopeAddress(from string) string {
	addr := strings.TrimSpace(from)
	if parsed, err := mail.ParseAddress(from); err == nil && parsed.Address != "" {
		return parsed.Address
	}
	if start := strings.LastIndex(addr, "<"); start >= 0 {
		if end := strings.LastIndex(addr, ">"); end > start {
			candidate := strings.TrimSpace(addr[start+1 : end])
			if candidate != "" {
				return candidate
			}
		}
	}
	return addr
}

func (s *NotificationService) rememberAppointment(event rabbitmq.AppointmentBookedEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appointments[event.AppointmentID] = event
}

func (s *NotificationService) getAppointment(appointmentID string) (rabbitmq.AppointmentBookedEvent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	appt, ok := s.appointments[appointmentID]
	return appt, ok
}

func (s *NotificationService) storePendingPayment(event rabbitmq.PaymentCompletedEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingPayments[event.AppointmentID] = event
}

func (s *NotificationService) consumePendingPayment(appointmentID string) (rabbitmq.PaymentCompletedEvent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.pendingPayments[appointmentID]
	if ok {
		delete(s.pendingPayments, appointmentID)
	}
	return event, ok
}

func (s *NotificationService) renderAppointmentConfirmationEmail(appt rabbitmq.AppointmentBookedEvent, payment rabbitmq.PaymentCompletedEvent, scheduledAt string) string {
	safeAppointmentID := html.EscapeString(appt.AppointmentID)
	safeScheduledAt := html.EscapeString(scheduledAt)
	safeConsultFee := html.EscapeString(fmt.Sprintf("%.2f", appt.ConsultFee))
	safePaymentID := html.EscapeString(payment.PaymentID)
	safePatientName := html.EscapeString(appt.PatientName)
	if safePatientName == "" {
		safePatientName = "Valued Patient"
	}
	safeDoctorName := html.EscapeString(appt.DoctorName)
	if safeDoctorName != "" && !strings.HasPrefix(strings.ToLower(safeDoctorName), "dr.") {
		safeDoctorName = "Dr. " + safeDoctorName
	}
	if safeDoctorName == "" {
		safeDoctorName = "Your practitioner"
	}

	mode := strings.TrimSpace(appt.ConsultationMode)
	if mode == "" {
		mode = "physical"
	}
	modeTitle := "In-Person Consultation"
	if strings.EqualFold(mode, "jitsi") || strings.EqualFold(mode, "telemedicine") {
		modeTitle = "Video Consultation"
	}

	// Keep the appointment confirmation clean for physical visits.
	meetingSection := ""
	if strings.EqualFold(mode, "jitsi") || strings.EqualFold(mode, "telemedicine") {
		meetingSection = fmt.Sprintf(`
    <div style="background-color:#f9fafb;border-radius:20px;padding:24px;margin-bottom:48px;">
      <div style="font-size:13px;color:#4b5563;line-height:1.7;">
        <strong style="color:#111827;">Meeting Information:</strong><br>
        For security reasons, we do not send session links via email. Please sign in to your <strong style="color:#0ea5e9;">Patient Dashboard</strong> at the scheduled time to access the video room or viewing instructions.
      </div>
    </div>`)
	}

	financialTitle := "Financial Details"
	if payment.PaymentID == "" {
		financialTitle = "Estimated Fee"
	}

	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <title>Appointment Confirmed</title>
</head>
<body style="margin:0;padding:60px 20px;background-color:#fcfcfd;font-family:'Inter', -apple-system, sans-serif;color:#1a1c1e;line-height:1.6;">
  <div style="max-width:540px;margin:0 auto;background-color:#ffffff;border:1px solid #f1f3f5;border-radius:32px;padding:48px;box-shadow:0 10px 40px rgba(0,0,0,0.02);">
    <header style="margin-bottom:48px;text-align:center;">
      <div style="width:48px;height:48px;background-color:#0ea5e9;border-radius:14px;display:inline-flex;margin-bottom:24px;align-items:center;justify-content:center;color:#ffffff;font-size:24px;font-weight:700;">H</div>
      <h1 style="font-size:28px;font-weight:700;letter-spacing:-0.03em;color:#111827;margin:0;">Booking Confirmed</h1>
      <p style="font-size:15px;color:#6b7280;margin-top:10px;">Hello %s, we've secured your appointment.</p>
    </header>

    <div style="margin-bottom:40px;border-top:1px solid #f1f3f5;padding-top:40px;">
      <table role="presentation" width="100%%" cellspacing="0" cellpadding="0">
        <tr>
          <td style="padding-bottom:30px;">
            <div style="font-size:11px;font-weight:700;color:#9ca3af;text-transform:uppercase;letter-spacing:0.1em;margin-bottom:8px;">Date & Time</div>
            <div style="font-size:16px;font-weight:600;color:#111827;">%s</div>
          </td>
        </tr>
        <tr>
          <td style="padding-bottom:30px;">
            <div style="font-size:11px;font-weight:700;color:#9ca3af;text-transform:uppercase;letter-spacing:0.1em;margin-bottom:8px;">Practitioner</div>
            <div style="font-size:16px;font-weight:600;color:#111827;">%s</div>
          </td>
        </tr>
        <tr>
          <td style="padding-bottom:10px;">
            <div style="font-size:11px;font-weight:700;color:#9ca3af;text-transform:uppercase;letter-spacing:0.1em;margin-bottom:8px;">Consultation Mode</div>
            <div style="font-size:16px;font-weight:600;color:#111827;">%s</div>
          </td>
        </tr>
      </table>
    </div>

    %s

    <div style="margin-bottom:40px;">
      <div style="font-size:11px;font-weight:700;color:#9ca3af;text-transform:uppercase;letter-spacing:0.1em;border-bottom:1px solid #f1f3f5;padding-bottom:12px;margin-bottom:20px;">%s</div>
      <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="font-size:14px;">
        <tr>
          <td style="color:#6b7280;padding-bottom:8px;">Consultation Fee</td>
          <td align="right" style="font-weight:600;color:#111827;">LKR %s</td>
        </tr>
        <tr>
          <td style="color:#9ca3af;font-size:12px;">Transaction Reference</td>
          <td align="right" style="font-family:monospace;color:#9ca3af;font-size:12px;">%s</td>
        </tr>
      </table>
    </div>

    <footer style="padding-top:40px;border-top:1px solid #f1f3f5;text-align:center;">
      <p style="font-size:12px;color:#9ca3af;margin:0;">Ref: %s</p>
      <p style="font-size:12px;color:#9ca3af;margin-top:4px;">&copy; 2026 Healthcare Platform. Local Office.</p>
    </footer>
  </div>
</body>
</html>`,
		safePatientName,
		safeScheduledAt,
		safeDoctorName,
		modeTitle,
		meetingSection,
		financialTitle,
		safeConsultFee,
		safePaymentID,
		safeAppointmentID,
	)
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
