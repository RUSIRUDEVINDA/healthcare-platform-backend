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
	"healthcare-platform/services/telemedicine-service/internal/model"
	"healthcare-platform/services/telemedicine-service/internal/repository"
)

type SessionService struct {
	repo                  *repository.SessionRepository
	log                   *logger.Logger
	appointmentServiceURL string
	jitsiBaseURL          string
	httpClient            *http.Client
}

func NewSessionService(repo *repository.SessionRepository, log *logger.Logger, appointmentServiceURL, jitsiBaseURL string) *SessionService {
	return &SessionService{
		repo:                  repo,
		log:                   log,
		appointmentServiceURL: strings.TrimRight(appointmentServiceURL, "/"),
		jitsiBaseURL:          strings.TrimRight(jitsiBaseURL, "/"),
		httpClient:            &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SessionService) CreateSession(callerID, callerRole, callerToken, appointmentID string, req *model.CreateSessionRequest) (*model.Session, error) {
	if callerRole != "patient" && callerRole != "doctor" {
		return nil, fmt.Errorf("only the assigned patient or doctor can create telemedicine sessions")
	}

	appt, err := s.fetchAppointment(callerToken, appointmentID)
	if err != nil {
		return nil, err
	}
	if !s.canAccessAppointment(appt, callerID, callerRole) {
		return nil, fmt.Errorf("forbidden: not your appointment")
	}
	if strings.EqualFold(appt.Status, "cancelled") {
		return nil, fmt.Errorf("appointment is cancelled")
	}

	if !strings.EqualFold(strings.TrimSpace(appt.ConsultationMode), string(model.ConsultationModeJitsi)) {
		return nil, fmt.Errorf("telemedicine sessions are only available for jitsi appointments")
	}

	roomName := strings.TrimSpace(appt.RoomName)
	joinURL := strings.TrimSpace(appt.JoinURL)
	if roomName == "" {
		var genErr error
		roomName, genErr = generateRoomName()
		if genErr != nil {
			return nil, fmt.Errorf("generate room name: %w", genErr)
		}
	}
	if joinURL == "" {
		joinURL = s.joinURL(roomName)
	}

	existing, err := s.repo.FindByRoomName(roomName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	session := &model.Session{
		AppointmentID:     appt.ID,
		PatientID:         appt.PatientID,
		DoctorID:          appt.DoctorID,
		DoctorOwnerUserID: appt.DoctorOwnerUserID,
		RoomName:          roomName,
		JoinURL:           joinURL,
		Provider:          "jitsi",
		Status:            model.SessionScheduled,
		Purpose:           strings.TrimSpace(req.Purpose),
		Notes:             strings.TrimSpace(req.Notes),
		CreatedByUserID:   callerID,
		CreatedByRole:     callerRole,
	}

	if err := s.repo.Create(session); err != nil {
		return nil, fmt.Errorf("service.CreateSession: %w", err)
	}

	s.log.Info("Telemedicine session created", "session_id", session.ID, "appointment_id", session.AppointmentID, "room", session.RoomName)
	return session, nil
}

func (s *SessionService) ListSessions(appointmentID, callerID, callerRole, callerToken string) ([]model.Session, error) {
	appt, err := s.fetchAppointment(callerToken, appointmentID)
	if err != nil {
		return nil, err
	}
	if !s.canAccessAppointment(appt, callerID, callerRole) {
		return nil, fmt.Errorf("forbidden: not your appointment")
	}

	return s.repo.ListByAppointment(appt.ID)
}

func (s *SessionService) GetSession(sessionID, callerID, callerRole, callerToken string) (*model.Session, error) {
	session, err := s.repo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, err
	}

	appt, err := s.fetchAppointment(callerToken, session.AppointmentID)
	if err != nil {
		return nil, err
	}
	if !s.canAccessAppointment(appt, callerID, callerRole) {
		return nil, fmt.Errorf("forbidden: not your appointment")
	}

	return session, nil
}

func (s *SessionService) GetJoinInfo(sessionID, callerID, callerRole, callerToken string) (*model.SessionJoinResponse, error) {
	session, err := s.GetSession(sessionID, callerID, callerRole, callerToken)
	if err != nil {
		return nil, err
	}

	return &model.SessionJoinResponse{
		SessionID:     session.ID,
		AppointmentID: session.AppointmentID,
		RoomName:      session.RoomName,
		JoinURL:       session.JoinURL,
		Provider:      session.Provider,
		Status:        session.Status,
		SessionNumber: session.SessionNumber,
	}, nil
}

func (s *SessionService) StartSession(sessionID, callerID, callerRole, callerToken string) (*model.Session, error) {
	session, err := s.GetSession(sessionID, callerID, callerRole, callerToken)
	if err != nil {
		return nil, err
	}
	if session.Status == model.SessionEnded {
		return nil, fmt.Errorf("session has already ended")
	}

	if err := s.repo.MarkStarted(sessionID); err != nil {
		return nil, err
	}

	return s.repo.GetByID(sessionID)
}

func (s *SessionService) EndSession(sessionID, callerID, callerRole, callerToken string) (*model.Session, error) {
	_, err := s.GetSession(sessionID, callerID, callerRole, callerToken)
	if err != nil {
		return nil, err
	}

	if err := s.repo.MarkEnded(sessionID); err != nil {
		return nil, err
	}

	return s.repo.GetByID(sessionID)
}

type appointmentResponse struct {
	ID                string `json:"id"`
	PatientID         string `json:"patient_id"`
	DoctorID          string `json:"doctor_id"`
	DoctorOwnerUserID string `json:"doctor_owner_user_id"`
	ConsultationMode  string `json:"consultation_mode"`
	RoomName          string `json:"room_name"`
	JoinURL           string `json:"join_url"`
	Status            string `json:"status"`
}

func (s *SessionService) fetchAppointment(token, appointmentID string) (*appointmentResponse, error) {
	req, err := http.NewRequest(http.MethodGet, s.appointmentServiceURL+"/api/v1/appointments/"+appointmentID, nil)
	if err != nil {
		return nil, fmt.Errorf("build appointment request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appointment service unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("appointment not found")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("not authorized")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden: not your appointment")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load appointment")
	}

	var appt appointmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&appt); err != nil {
		return nil, fmt.Errorf("invalid appointment service response")
	}
	if appt.ID == "" {
		return nil, fmt.Errorf("appointment not found")
	}

	return &appt, nil
}

func (s *SessionService) canAccessAppointment(appt *appointmentResponse, callerID, callerRole string) bool {
	switch callerRole {
	case "patient":
		return appt.PatientID == callerID
	case "doctor":
		return appt.DoctorOwnerUserID == callerID
	default:
		return false
	}
}

func (s *SessionService) joinURL(roomName string) string {
	return s.jitsiBaseURL + "/" + url.PathEscape(roomName)
}

func (s *SessionService) CreateFromAppointmentBooked(event rabbitmq.AppointmentBookedEvent) error {
	if !strings.EqualFold(strings.TrimSpace(event.ConsultationMode), string(model.ConsultationModeJitsi)) {
		return nil
	}

	roomName := strings.TrimSpace(event.RoomName)
	joinURL := strings.TrimSpace(event.JoinURL)
	if roomName == "" {
		var err error
		roomName, err = generateRoomName()
		if err != nil {
			return fmt.Errorf("generate room name: %w", err)
		}
	}
	if joinURL == "" {
		joinURL = s.joinURL(roomName)
	}

	existing, err := s.repo.FindByRoomName(roomName)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	session := &model.Session{
		AppointmentID:     event.AppointmentID,
		PatientID:         event.PatientID,
		DoctorID:          event.DoctorID,
		DoctorOwnerUserID: event.DoctorOwnerUserID,
		RoomName:          roomName,
		JoinURL:           joinURL,
		Provider:          "jitsi",
		Status:            model.SessionScheduled,
		CreatedByUserID:   event.PatientID,
		CreatedByRole:     "patient",
	}

	return s.repo.Create(session)
}

func (s *SessionService) DeletePatientSessions(patientID string) error {
	if err := s.repo.DeleteByPatientID(patientID); err != nil {
		return fmt.Errorf("service.DeletePatientSessions: %w", err)
	}
	s.log.Info("Patient telemedicine sessions deleted", "patient_id", patientID)
	return nil
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
