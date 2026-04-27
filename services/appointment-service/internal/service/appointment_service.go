package service

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/pkg/rabbitmq"
	"healthcare-platform/services/appointment-service/internal/model"
	"healthcare-platform/services/appointment-service/internal/repository"
)

type AppointmentService struct {
	repo              *repository.AppointmentRepository
	mq                *rabbitmq.Client
	log               *logger.Logger
	doctorServiceURL  string
	patientServiceURL string
	internalAPIKey    string
	jitsiBaseURL      string
	httpClient        *http.Client
}

const hospitalFee = 500.0

func NewAppointmentService(
	repo *repository.AppointmentRepository,
	mq *rabbitmq.Client,
	log *logger.Logger,
	doctorServiceURL,
	patientServiceURL,
	internalAPIKey,
	jitsiBaseURL string,
) *AppointmentService {
	return &AppointmentService{
		repo:              repo,
		mq:                mq,
		log:               log,
		doctorServiceURL:  strings.TrimRight(doctorServiceURL, "/"),
		patientServiceURL: strings.TrimRight(patientServiceURL, "/"),
		internalAPIKey:    strings.TrimSpace(internalAPIKey),
		jitsiBaseURL:      strings.TrimRight(jitsiBaseURL, "/"),
		httpClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *AppointmentService) DeletePatientAppointments(patientID string) error {
	s.log.Info("Deleting all appointments for patient", "patient_id", patientID)
	return s.repo.DeleteByPatientID(patientID)
}

func (s *AppointmentService) BookAppointment(patientID, role, callerToken, patientEmail, firstName, lastName string, req *model.BookAppointmentRequest) (*model.Appointment, error) {
	if role != "patient" {
		return nil, fmt.Errorf("only patients can book appointments")
	}

	paymentMode := req.PaymentMode
	if paymentMode == "" {
		paymentMode = model.PaymentModePayNow
	}

	consultationMode := req.ConsultationMode
	if consultationMode == "" {
		consultationMode = model.ConsultationModePhysical
	}
	if !model.IsValidConsultationMode(consultationMode) {
		return nil, fmt.Errorf("invalid consultation mode; allowed values: physical, jitsi")
	}
	if paymentMode == model.PaymentModePayNow && !req.PaymentCompleted {
		return nil, fmt.Errorf("pay now appointments must be finalized after payment")
	}
	if paymentMode == model.PaymentModePayLater && req.PaymentCompleted {
		return nil, fmt.Errorf("pay later appointments cannot be finalized as paid")
	}

	now := time.Now().UTC()
	appt := &model.Appointment{
		PatientID:        patientID,
		PatientFirstName: firstName,
		PatientLastName:  lastName,
		Notes:            req.Notes,
		ConsultationMode: consultationMode,
	}
	if strings.TrimSpace(req.AppointmentID) != "" {
		appt.ID = strings.TrimSpace(req.AppointmentID)
	}

	if strings.TrimSpace(req.SlotID) != "" {
		slot, err := s.repo.GetSlotByID(req.SlotID)
		if err != nil {
			return nil, fmt.Errorf("service.BookAppointment load slot: %w", err)
		}
		if slot == nil {
			return nil, fmt.Errorf("slot not found")
		}
		if slot.IsBooked {
			return nil, fmt.Errorf("slot is already booked")
		}
		if !slot.StartTime.After(now) {
			return nil, fmt.Errorf("scheduled time must be in the future")
		}
		if paymentMode == model.PaymentModePayLater && time.Until(slot.StartTime) <= time.Hour {
			return nil, fmt.Errorf("pay later is not allowed when the appointment starts within one hour")
		}

		dueAt := slot.StartTime.UTC()
		if oneHourFromNow := now.Add(time.Hour); oneHourFromNow.Before(dueAt) {
			dueAt = oneHourFromNow
		}

		appt.SlotID = slot.ID
		appt.DoctorID = slot.DoctorID
		appt.DoctorOwnerUserID = slot.OwnerUserID
		appt.ScheduledAt = slot.StartTime.UTC()
		appt.DurationMinutes = int(slot.EndTime.Sub(slot.StartTime).Minutes())
		appt.PaymentDueAt = &dueAt
	} else {
		if req.ScheduledAt == nil || req.DurationMinutes == nil || strings.TrimSpace(req.DoctorID) == "" {
			return nil, fmt.Errorf("either slot_id or doctor_id, scheduled_at, and duration_minutes must be provided")
		}
		if !req.ScheduledAt.After(now) {
			return nil, fmt.Errorf("scheduled time must be in the future")
		}
		if paymentMode == model.PaymentModePayLater && time.Until(req.ScheduledAt.UTC()) <= time.Hour {
			return nil, fmt.Errorf("pay later is not allowed when the appointment starts within one hour")
		}

		dueAt := req.ScheduledAt.UTC()
		oneHourFromNow := now.Add(time.Hour)
		if oneHourFromNow.Before(dueAt) {
			dueAt = oneHourFromNow
		}

		appt.DoctorID = req.DoctorID
		appt.ScheduledAt = req.ScheduledAt.UTC()
		appt.DurationMinutes = *req.DurationMinutes
		appt.PaymentDueAt = &dueAt
	}

	if consultationMode == model.ConsultationModeJitsi {
		roomName, err := generateRoomName()
		if err != nil {
			return nil, fmt.Errorf("generate room name: %w", err)
		}
		appt.RoomName = roomName
		appt.JoinURL = s.joinURL(roomName)
	}

	if req.PaymentCompleted {
		appt.PaymentStatus = model.PaymentPaid
		paidAt := now
		appt.PaidAt = &paidAt
		appt.PaymentDueAt = nil
	} else {
		appt.PaymentStatus = model.PaymentPending
	}

	doctorProfile, err := s.getDoctorProfile("", "/doctors/"+url.PathEscape(strings.TrimSpace(appt.DoctorID)))
	if err != nil {
		return nil, err
	}
	doctorFee := doctorProfile.Data.ChannelingFee
	totalConsultFee := doctorFee + hospitalFee
	appt.ConsultFee = totalConsultFee

	if err := s.repo.Create(appt); err != nil {
		return nil, fmt.Errorf("service.BookAppointment: %w", err)
	}

	event := rabbitmq.AppointmentBookedEvent{
		AppointmentID:     appt.ID,
		PatientID:         appt.PatientID,
		PatientName:       fmt.Sprintf("%s %s", appt.PatientFirstName, appt.PatientLastName),
		DoctorID:          appt.DoctorID,
		DoctorName:        strings.TrimSpace(string(doctorProfile.Data.Name)),
		DoctorOwnerUserID: appt.DoctorOwnerUserID,
		PaymentStatus:     string(appt.PaymentStatus),
		ConsultationMode:  string(appt.ConsultationMode),
		RoomName:          appt.RoomName,
		JoinURL:           appt.JoinURL,
		PatientEmail:      strings.TrimSpace(patientEmail),
		ScheduledAt:       appt.ScheduledAt.Format(time.RFC3339),
		ConsultFee:        totalConsultFee,
	}
	// Clean up doctor name if it was JSON string
	event.DoctorName = strings.Trim(event.DoctorName, `"`)
	if err := s.mq.PublishAppointmentBooked(event); err != nil {
		s.log.Error("Failed to publish appointment.booked event", "error", err)
	}

	s.log.Info("Appointment booked successfully", "id", appt.ID, "payment_status", appt.PaymentStatus)
	return s.sanitizeAppointmentForResponse(appt), nil
}

func (s *AppointmentService) GetStatus(id, callerID, role string) (*model.Appointment, error) {
	appt, err := s.repo.GetByID(id)
	if err != nil || appt == nil {
		return nil, err
	}
	if !s.canAccessAppointment(appt, callerID, role) {
		return nil, fmt.Errorf("forbidden: not your appointment")
	}
	return s.sanitizeAppointmentForResponse(appt), nil
}

func (s *AppointmentService) ListAppointments(callerID, role string) ([]model.Appointment, error) {
	var appts []model.Appointment
	var err error
	switch role {
	case "patient":
		appts, err = s.repo.ListAppointmentsByPatient(callerID)
	case "doctor":
		appts, err = s.repo.ListAppointmentsByDoctorOwner(callerID)
	case "admin":
		appts, err = s.repo.ListAppointmentsAll()
	default:
		return nil, fmt.Errorf("forbidden: invalid role")
	}
	if err != nil {
		return nil, err
	}
	s.enrichMissingPatientNames(appts)

	for i := range appts {
		appts[i] = *s.sanitizeAppointmentForResponse(&appts[i])
	}
	return appts, nil
}

func (s *AppointmentService) enrichMissingPatientNames(appts []model.Appointment) {
	if len(appts) == 0 || s.patientServiceURL == "" || s.internalAPIKey == "" {
		return
	}

	cache := make(map[string][2]string, len(appts))
	for i := range appts {
		if strings.TrimSpace(appts[i].PatientFirstName) != "" || strings.TrimSpace(appts[i].PatientLastName) != "" {
			continue
		}
		patientID := strings.TrimSpace(appts[i].PatientID)
		if patientID == "" {
			continue
		}

		names, ok := cache[patientID]
		if !ok {
			first, last, err := s.fetchPatientDisplayName(patientID)
			if err != nil {
				s.log.Warn("Failed to enrich patient display name", "patient_id", patientID, "error", err)
				cache[patientID] = [2]string{}
				continue
			}
			names = [2]string{first, last}
			cache[patientID] = names
		}

		if names[0] != "" || names[1] != "" {
			appts[i].PatientFirstName = names[0]
			appts[i].PatientLastName = names[1]
		}
	}
}

func (s *AppointmentService) fetchPatientDisplayName(patientUserID string) (string, string, error) {
	endpoint := fmt.Sprintf("%s/internal/v1/patients/by-user/%s", s.patientServiceURL, url.PathEscape(patientUserID))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Internal-Api-Key", s.internalAPIKey)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("request patient-service: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "", "", nil
	}
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	var payload struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}
	return strings.TrimSpace(payload.FirstName), strings.TrimSpace(payload.LastName), nil
}

func (s *AppointmentService) CancelAppointment(id, callerID, role string) error {
	appt, err := s.repo.GetByID(id)
	if err != nil || appt == nil {
		return fmt.Errorf("appointment not found")
	}
	if !s.canManageAppointment(appt, callerID, role) {
		return fmt.Errorf("forbidden: not your appointment")
	}

	paymentStatus := appt.PaymentStatus
	if paymentStatus != model.PaymentPaid {
		paymentStatus = model.PaymentExpired
	}

	if err := s.repo.CancelAndRelease(id, paymentStatus); err != nil {
		return err
	}

	event := rabbitmq.AppointmentCancelledEvent{
		AppointmentID: appt.ID,
		PatientID:     appt.PatientID,
		DoctorID:      appt.DoctorID,
		CancelledBy:   role,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
	if err := s.mq.PublishAppointmentCancelled(event); err != nil {
		s.log.Error("Failed to publish appointment.cancelled event", "error", err)
	}

	return nil
}

func (s *AppointmentService) GetDoctorSlots(doctorID, status string) ([]model.Slot, error) {
	filter := strings.ToLower(strings.TrimSpace(status))
	switch filter {
	case "", "all":
		filter = "all"
	case "available", "booked":
	default:
		return nil, fmt.Errorf("invalid slot status filter")
	}

	return s.repo.GetSlotsByDoctor(doctorID, filter)
}

func (s *AppointmentService) CreateSlot(callerID, callerToken, role string, req *model.CreateSlotRequest) (*model.Slot, error) {
	if role != "doctor" && role != "admin" {
		return nil, fmt.Errorf("only doctors or admins can create slots")
	}
	if !req.EndTime.After(req.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	doctorProfileID := strings.TrimSpace(req.DoctorID)
	if role == "doctor" {
		resolvedID, err := s.getCallerDoctorProfileID(callerToken)
		if err != nil {
			return nil, err
		}
		if doctorProfileID != "" && doctorProfileID != resolvedID {
			return nil, fmt.Errorf("forbidden: you can only create slots for your own doctor profile")
		}
		doctorProfileID = resolvedID
	}
	if doctorProfileID == "" {
		return nil, fmt.Errorf("doctor_id is required")
	}
	hospital := strings.TrimSpace(req.Hospital)
	if hospital == "" {
		return nil, fmt.Errorf("hospital is required")
	}

	slot := &model.Slot{
		DoctorID:    doctorProfileID,
		OwnerUserID: callerID,
		Hospital:    hospital,
		StartTime:   req.StartTime.UTC(),
		EndTime:     req.EndTime.UTC(),
		IsBooked:    false,
	}

	if err := s.repo.CreateSlot(slot); err != nil {
		return nil, fmt.Errorf("service.CreateSlot: %w", err)
	}

	return slot, nil
}

func (s *AppointmentService) UpdateSlot(id, callerID, role string, req *model.UpdateSlotRequest) (*model.Slot, error) {
	if role != "doctor" && role != "admin" {
		return nil, fmt.Errorf("only doctors or admins can update slots")
	}

	slot, err := s.repo.GetSlotByID(id)
	if err != nil || slot == nil {
		return nil, fmt.Errorf("slot not found")
	}
	if role != "admin" && slot.OwnerUserID != callerID {
		return nil, fmt.Errorf("forbidden: not your slot")
	}

	if req.IsBooked != nil {
		return nil, fmt.Errorf("slot booking state is managed by the system")
	}
	if req.StartTime != nil {
		slot.StartTime = req.StartTime.UTC()
	}
	if req.EndTime != nil {
		slot.EndTime = req.EndTime.UTC()
	}
	if req.Hospital != nil {
		slot.Hospital = strings.TrimSpace(*req.Hospital)
		if slot.Hospital == "" {
			return nil, fmt.Errorf("hospital is required")
		}
	}
	if !slot.EndTime.After(slot.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	if err := s.repo.UpdateSlot(slot); err != nil {
		return nil, fmt.Errorf("service.UpdateSlot: %w", err)
	}

	return slot, nil
}

func (s *AppointmentService) DeleteSlot(id, callerID, role string) error {
	if role != "doctor" && role != "admin" {
		return fmt.Errorf("only doctors or admins can delete slots")
	}

	slot, err := s.repo.GetSlotByID(id)
	if err != nil || slot == nil {
		return fmt.Errorf("slot not found")
	}
	if role != "admin" && slot.OwnerUserID != callerID {
		return fmt.Errorf("forbidden: not your slot")
	}
	if slot.IsBooked {
		return fmt.Errorf("cannot delete a booked slot")
	}

	return s.repo.DeleteSlot(id)
}

func (s *AppointmentService) ListAllSlots(callerID, role string) ([]model.Slot, error) {
	switch role {
	case "admin":
		return s.repo.ListSlots()
	case "doctor":
		return s.repo.ListSlotsByOwner(callerID)
	default:
		return nil, fmt.Errorf("forbidden: only doctors or admins can list slots")
	}
}

func (s *AppointmentService) UpdateAppointmentStatus(id, callerID, role string, status model.AppointmentStatus) error {
	if role != "doctor" && role != "admin" {
		return fmt.Errorf("only doctors or admins can update appointment status")
	}

	appt, err := s.repo.GetByID(id)
	if err != nil || appt == nil {
		return fmt.Errorf("appointment not found")
	}
	if role != "admin" && appt.DoctorOwnerUserID != callerID {
		return fmt.Errorf("forbidden: not your appointment")
	}

	return s.repo.UpdateStatus(id, status)
}

func (s *AppointmentService) HandlePaymentCompleted(appointmentID string) error {
	appt, err := s.repo.GetByID(appointmentID)
	if err != nil {
		return fmt.Errorf("appointment not found")
	}
	if appt == nil {
		s.log.Info("Payment completed before appointment creation; skipping until appointment exists", "appointment_id", appointmentID)
		return nil
	}
	if appt.Status == model.StatusCancelled {
		return nil
	}
	if appt.PaymentStatus == model.PaymentPaid {
		return nil
	}

	if err := s.repo.MarkPaymentCompleted(appointmentID); err != nil {
		return fmt.Errorf("service.HandlePaymentCompleted: %w", err)
	}

	s.log.Info("Appointment marked as paid", "appointment_id", appointmentID)
	return nil
}

func (s *AppointmentService) HandlePaymentRefunded(appointmentID string) error {
	return s.repo.UpdatePaymentStatus(appointmentID, model.PaymentRefunded)
}

func (s *AppointmentService) sanitizeAppointmentForResponse(appt *model.Appointment) *model.Appointment {
	if appt == nil {
		return nil
	}

	sanitized := *appt
	if sanitized.PaymentStatus != model.PaymentPaid {
		sanitized.JoinURL = ""
	}
	return &sanitized
}

func (s *AppointmentService) ExpireOverdueAppointments(now time.Time) error {
	overdue, err := s.repo.FindOverdueUnpaid(now.UTC())
	if err != nil {
		return err
	}

	for _, appt := range overdue {
		if err := s.repo.CancelAndRelease(appt.ID, model.PaymentExpired); err != nil {
			s.log.Error("Failed to expire appointment", "appointment_id", appt.ID, "error", err)
		}
	}

	if len(overdue) > 0 {
		s.log.Info("Expired unpaid appointments", "count", len(overdue))
	}
	return nil
}

func (s *AppointmentService) canAccessAppointment(appt *model.Appointment, callerID, role string) bool {
	switch role {
	case "admin":
		return true
	case "doctor":
		return appt.DoctorOwnerUserID == callerID
	case "patient":
		return appt.PatientID == callerID
	default:
		return false
	}
}

func (s *AppointmentService) canManageAppointment(appt *model.Appointment, callerID, role string) bool {
	switch role {
	case "admin":
		return true
	case "doctor":
		return appt.DoctorOwnerUserID == callerID
	case "patient":
		return appt.PatientID == callerID
	default:
		return false
	}
}

type doctorProfileResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID            json.RawMessage `json:"id"`
		Name          json.RawMessage `json:"name"`
		ChannelingFee float64         `json:"channeling_fee"`
	} `json:"data"`
}

func (s *AppointmentService) getCallerDoctorProfileID(token string) (string, error) {
	profile, err := s.getDoctorProfile(token, "/doctors/me")
	if err != nil {
		return "", err
	}

	id := strings.TrimSpace(string(profile.Data.ID))
	id = strings.Trim(id, `"`)
	if id == "" {
		return "", fmt.Errorf("doctor profile not found for current user")
	}
	return id, nil
}

func (s *AppointmentService) getDoctorConsultationFee(doctorID string) (float64, error) {
	profile, err := s.getDoctorProfile("", "/doctors/"+url.PathEscape(strings.TrimSpace(doctorID)))
	if err != nil {
		return 0, err
	}
	if profile.Data.ChannelingFee < 0 {
		return 0, fmt.Errorf("doctor channeling fee is invalid")
	}
	if profile.Data.ChannelingFee == 0 {
		s.log.Warn("Doctor channeling fee is zero; proceeding with hospital fee only", "doctor_id", doctorID)
	}
	return profile.Data.ChannelingFee, nil
}

func (s *AppointmentService) getDoctorProfile(token, path string) (*doctorProfileResponse, error) {
	req, err := http.NewRequest(http.MethodGet, s.doctorServiceURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build doctor profile request: %w", err)
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doctor service unavailable")
	}
	defer resp.Body.Close()

	var body doctorProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("invalid doctor service response")
	}
	if !body.Success || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("doctor profile not found")
	}
	return &body, nil
}

func (s *AppointmentService) joinURL(roomName string) string {
	return s.jitsiBaseURL + "/" + url.PathEscape(roomName)
}

func generateRoomName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	encoded = strings.ToLower(encoded)
	return "telemed-" + encoded, nil
}
