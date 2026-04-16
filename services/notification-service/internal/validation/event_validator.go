package validation

import (
	"fmt"
	"strings"

	"healthcare-platform/pkg/rabbitmq"
)

func ValidateAppointmentBookedEventLenient(event *rabbitmq.AppointmentBookedEvent) error {
	if event == nil {
		return fmt.Errorf("appointment.booked event is nil")
	}

	event.AppointmentID = strings.TrimSpace(event.AppointmentID)
	event.DoctorID = strings.TrimSpace(event.DoctorID)
	event.PatientID = strings.TrimSpace(event.PatientID)
	event.ScheduledAt = strings.TrimSpace(event.ScheduledAt)
	event.Time = strings.TrimSpace(event.Time)

	if event.ScheduledAt == "" && event.Time != "" {
		event.ScheduledAt = event.Time
	}
	if event.Time == "" && event.ScheduledAt != "" {
		event.Time = event.ScheduledAt
	}

	if event.AppointmentID == "" {
		return fmt.Errorf("appointmentId is required (empty string)")
	}
	if event.DoctorID == "" {
		return fmt.Errorf("doctorId is required (empty string)")
	}
	if event.PatientID == "" {
		return fmt.Errorf("patientId is required (empty string)")
	}
	if event.ScheduledAt == "" {
		return fmt.Errorf("scheduled_at or time is required (both empty)")
	}

	return nil
}
