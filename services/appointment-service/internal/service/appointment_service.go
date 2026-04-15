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
	jitsiBaseURL      string
	patientServiceURL string
	internalAPIKey    string
	httpClient        *http.Client
}

func NewAppointmentService(repo *repository.AppointmentRepository, mq *rabbitmq.Client, log *logger.Logger, doctorServiceURL, jitsiBaseURL, patientServiceURL, internalAPIKey string) *AppointmentService {
	return &AppointmentService{
		repo:              repo,
		mq:                mq,
		log:               log,
		doctorServiceURL:  strings.TrimRight(doctorServiceURL, "/"),
		jitsiBaseURL:      strings.TrimRight(jitsiBaseURL, "/"),
		patientServiceURL: strings.TrimRight(patientServiceURL, "/"),
		internalAPIKey:    internalAPIKey,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *AppointmentService) BookAppointment(patientID, role, callerToken, patientFirstName, patientLastName string, req *model.BookAppointmentRequest) (*model.Appointment, error) {
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

	now := time.Now().UTC()
	appt := &model.Appointment{
		PatientID:        patientID,
		PatientFirstName: strings.TrimSpace(patientFirstName),
		PatientLastName:  strings.TrimSpace(patientLastName),
		Notes:            req.Notes,
		ConsultationMode: consultationMode,
		PaymentStatus:    model.PaymentPending,
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

	if err := s.repo.Create(appt); err != nil {
		return nil, fmt.Errorf("service.BookAppointment: %w", err)
	}

	event := rabbitmq.AppointmentBookedEvent{
		AppointmentID:     appt.ID,
		PatientID:         appt.PatientID,
		DoctorID:          appt.DoctorID,
		DoctorOwnerUserID: appt.DoctorOwnerUserID,
		ConsultationMode:  string(appt.ConsultationMode),
		RoomName:          appt.RoomName,
		JoinURL:           appt.JoinURL,
		ScheduledAt:       appt.ScheduledAt.Format(time.RFC3339),
		ConsultFee:        1500.00,
	}
	if err := s.mq.PublishAppointmentBooked(event); err != nil {
		s.log.Error("Failed to publish appointment.booked event", "error", err)
	}

	s.log.Info("Appointment booked successfully", "id", appt.ID)
	return appt, nil
}

func (s *AppointmentService) GetStatus(id, callerID, role string) (*model.Appointment, error) {
	appt, err := s.repo.GetByID(id)
	if err != nil || appt == nil {
		return nil, err
	}
	if !s.canAccessAppointment(appt, callerID, role) {
		return nil, fmt.Errorf("forbidden: not your appointment")
	}
	if role == "doctor" {
		batch := []model.Appointment{*appt}
		s.enrichDoctorAppointmentPatientNames(batch)
		*appt = batch[0]
	}
	return appt, nil
}

func (s *AppointmentService) ListAppointments(callerID, role string) ([]model.Appointment, error) {
	switch role {
	case "patient":
		return s.repo.ListAppointmentsByPatient(callerID)
	case "doctor":
		appts, err := s.repo.ListAppointmentsByDoctorOwner(callerID)
		if err != nil {
			return nil, err
		}
		s.enrichDoctorAppointmentPatientNames(appts)
		return appts, nil
	case "admin":
		return s.repo.ListAppointmentsAll()
	default:
		return nil, fmt.Errorf("forbidden: invalid role")
	}
}

func (s *AppointmentService) enrichDoctorAppointmentPatientNames(appts []model.Appointment) {
	if len(appts) == 0 || s.patientServiceURL == "" || s.internalAPIKey == "" {
		return
	}
	need := make(map[string]struct{})
	for i := range appts {
		a := &appts[i]
		if strings.TrimSpace(a.PatientFirstName) != "" || strings.TrimSpace(a.PatientLastName) != "" {
			continue
		}
		if pid := strings.TrimSpace(a.PatientID); pid != "" {
			need[pid] = struct{}{}
		}
	}
	cache := make(map[string]struct{ fn, ln string })
	for pid := range need {
		fn, ln, ok := s.lookupPatientDisplayName(pid)
		if ok {
			cache[pid] = struct{ fn, ln string }{fn, ln}
		}
	}
	for i := range appts {
		a := &appts[i]
		if strings.TrimSpace(a.PatientFirstName) != "" || strings.TrimSpace(a.PatientLastName) != "" {
			continue
		}
		if c, ok := cache[strings.TrimSpace(a.PatientID)]; ok {
			a.PatientFirstName = c.fn
			a.PatientLastName = c.ln
		}
	}
}

func (s *AppointmentService) lookupPatientDisplayName(patientUserID string) (firstName, lastName string, ok bool) {
	u := fmt.Sprintf("%s/internal/v1/patients/by-user/%s", s.patientServiceURL, url.PathEscape(patientUserID))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", "", false
	}
	req.Header.Set("X-Internal-Api-Key", s.internalAPIKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Warn("patient-service lookup failed", "error", err)
		return "", "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", false
	}
	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", false
	}
	fn := strings.TrimSpace(body.FirstName)
	ln := strings.TrimSpace(body.LastName)
	if fn == "" && ln == "" {
		return "", "", false
	}
	return fn, ln, true
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

	now := time.Now().UTC()
	st := req.StartTime.UTC()
	et := req.EndTime.UTC()
	if st.Before(now) {
		return nil, fmt.Errorf("slot start time cannot be in the past")
	}
	if et.Before(now) {
		return nil, fmt.Errorf("slot end time cannot be in the past")
	}

	doctorProfileID := strings.TrimSpace(req.DoctorID)
	slotHospital := strings.TrimSpace(req.Hospital)

	if role == "doctor" {
		resolvedID, profileHospital, err := s.fetchCallerDoctorMe(callerToken)
		if err != nil {
			return nil, err
		}
		profHosp := strings.TrimSpace(profileHospital)
		if profHosp == "" {
			return nil, fmt.Errorf("add your hospital on your profile before creating slots")
		}
		if doctorProfileID != "" && doctorProfileID != resolvedID {
			return nil, fmt.Errorf("forbidden: you can only create slots for your own doctor profile")
		}
		doctorProfileID = resolvedID
		if slotHospital == "" {
			slotHospital = profHosp
		}
		if slotHospital != profHosp {
			return nil, fmt.Errorf("hospital must match your profile hospital (%s)", profHosp)
		}
	} else if role == "admin" {
		if doctorProfileID == "" {
			return nil, fmt.Errorf("doctor_id is required")
		}
	}

	slot := &model.Slot{
		DoctorID:    doctorProfileID,
		OwnerUserID: callerID,
		StartTime:   req.StartTime.UTC(),
		EndTime:     req.EndTime.UTC(),
		IsBooked:    false,
		Hospital:    slotHospital,
	}

	if err := s.repo.CreateSlot(slot); err != nil {
		return nil, fmt.Errorf("service.CreateSlot: %w", err)
	}

	return slot, nil
}

func (s *AppointmentService) UpdateSlot(id, callerID, callerToken, role string, req *model.UpdateSlotRequest) (*model.Slot, error) {
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
	if slot.IsBooked {
		return nil, fmt.Errorf("cannot update a booked slot")
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
		h := strings.TrimSpace(*req.Hospital)
		if role == "doctor" {
			_, profileHospital, derr := s.fetchCallerDoctorMe(callerToken)
			if derr != nil {
				return nil, derr
			}
			profHosp := strings.TrimSpace(profileHospital)
			if profHosp == "" {
				return nil, fmt.Errorf("add your hospital on your profile before updating slots")
			}
			if h == "" {
				h = profHosp
			}
			if h != profHosp {
				return nil, fmt.Errorf("hospital must match your profile hospital (%s)", profHosp)
			}
			slot.Hospital = h
		} else {
			slot.Hospital = h
		}
	}
	if !slot.EndTime.After(slot.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	now := time.Now().UTC()
	if slot.StartTime.UTC().Before(now) {
		return nil, fmt.Errorf("slot start time cannot be in the past")
	}
	if slot.EndTime.UTC().Before(now) {
		return nil, fmt.Errorf("slot end time cannot be in the past")
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
	if err != nil || appt == nil {
		return fmt.Errorf("appointment not found")
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

type doctorMePayload struct {
	ID       json.RawMessage `json:"id"`
	Hospital string          `json:"hospital"`
}

type doctorMeResponse struct {
	Success bool            `json:"success"`
	Data    doctorMePayload `json:"data"`
}

func normalizeDoctorProfileID(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	return strings.Trim(s, `"`)
}

func (s *AppointmentService) fetchCallerDoctorMe(token string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, s.doctorServiceURL+"/doctors/me", nil)
	if err != nil {
		return "", "", fmt.Errorf("build doctor profile request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("doctor service unavailable")
	}
	defer resp.Body.Close()

	var body doctorMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", fmt.Errorf("invalid doctor service response")
	}
	if !body.Success || len(body.Data.ID) == 0 || resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("doctor profile not found for current user")
	}

	id := normalizeDoctorProfileID(body.Data.ID)
	if id == "" {
		return "", "", fmt.Errorf("doctor profile not found for current user")
	}
	return id, body.Data.Hospital, nil
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
