package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"healthcare-platform/services/appointment-service/internal/model"
)

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(a *model.Appointment) error {
	a.ID = uuid.New().String()
	a.Status = model.StatusPending
	if a.PaymentStatus == "" {
		a.PaymentStatus = model.PaymentPending
	}
	a.CreatedAt = time.Now().UTC()
	a.UpdatedAt = time.Now().UTC()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var slotID string
	var ownerUserID string
	var slotStart time.Time
	var slotEnd time.Time

	if a.SlotID != "" {
		err = tx.QueryRow(`
			SELECT id, doctor_id, owner_user_id, start_time, end_time
			FROM slots
			WHERE id = $1
			  AND is_booked = FALSE
			FOR UPDATE SKIP LOCKED`,
			a.SlotID,
		).Scan(&slotID, &a.DoctorID, &ownerUserID, &slotStart, &slotEnd)
	} else {
		err = tx.QueryRow(`
			SELECT id, doctor_id, owner_user_id, start_time, end_time
			FROM slots
			WHERE doctor_id = $1
			  AND is_booked = FALSE
			  AND start_time <= $2
			  AND end_time >= $2
			ORDER BY start_time ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED`,
			a.DoctorID, a.ScheduledAt,
		).Scan(&slotID, &a.DoctorID, &ownerUserID, &slotStart, &slotEnd)
	}
	if err == sql.ErrNoRows {
		return fmt.Errorf("no available slot found for requested time")
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE slots SET is_booked = TRUE WHERE id = $1`, slotID); err != nil {
		return err
	}

	a.SlotID = slotID
	a.DoctorOwnerUserID = ownerUserID
	a.ScheduledAt = slotStart
	a.DurationMinutes = int(slotEnd.Sub(slotStart).Minutes())

	_, err = tx.Exec(`
		INSERT INTO appointments (
			id, patient_id, patient_first_name, patient_last_name, doctor_id, doctor_owner_user_id, slot_id, consultation_mode, room_name, join_url, scheduled_at,
			duration_minutes, status, payment_status, payment_due_at, paid_at,
			notes, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		a.ID, a.PatientID, a.PatientFirstName, a.PatientLastName, a.DoctorID, a.DoctorOwnerUserID, a.SlotID, a.ConsultationMode, a.RoomName, a.JoinURL, a.ScheduledAt,
		a.DurationMinutes, a.Status, a.PaymentStatus, a.PaymentDueAt, a.PaidAt,
		a.Notes, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AppointmentRepository) GetByID(id string) (*model.Appointment, error) {
	a := &model.Appointment{}
	var paymentDueAt sql.NullTime
	var paidAt sql.NullTime
	var slotID sql.NullString
	var roomName sql.NullString
	var joinURL sql.NullString

	err := r.db.QueryRow(`
		SELECT id, patient_id, patient_first_name, patient_last_name, doctor_id, doctor_owner_user_id, slot_id, consultation_mode, room_name, join_url, scheduled_at, duration_minutes,
		       status, payment_status, payment_due_at, paid_at, notes, created_at, updated_at
		FROM appointments WHERE id = $1`, id).
		Scan(
			&a.ID, &a.PatientID, &a.PatientFirstName, &a.PatientLastName, &a.DoctorID, &a.DoctorOwnerUserID, &slotID, &a.ConsultationMode, &roomName, &joinURL, &a.ScheduledAt, &a.DurationMinutes,
			&a.Status, &a.PaymentStatus, &paymentDueAt, &paidAt, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if paymentDueAt.Valid {
		t := paymentDueAt.Time
		a.PaymentDueAt = &t
	}
	if paidAt.Valid {
		t := paidAt.Time
		a.PaidAt = &t
	}
	if slotID.Valid {
		a.SlotID = slotID.String
	}
	if roomName.Valid {
		a.RoomName = roomName.String
	}
	if joinURL.Valid {
		a.JoinURL = joinURL.String
	}
	return a, nil
}

func (r *AppointmentRepository) ListAppointmentsByPatient(patientID string) ([]model.Appointment, error) {
	return r.listAppointments(`WHERE patient_id = $1 ORDER BY scheduled_at DESC`, patientID)
}

func (r *AppointmentRepository) ListAppointmentsByDoctorOwner(ownerUserID string) ([]model.Appointment, error) {
	return r.listAppointments(`WHERE doctor_owner_user_id = $1 ORDER BY scheduled_at DESC`, ownerUserID)
}

func (r *AppointmentRepository) ListAppointmentsAll() ([]model.Appointment, error) {
	return r.listAppointments(`ORDER BY scheduled_at DESC`)
}

func (r *AppointmentRepository) listAppointments(clause string, args ...interface{}) ([]model.Appointment, error) {
	query := `
		SELECT id, patient_id, patient_first_name, patient_last_name, doctor_id, doctor_owner_user_id, slot_id, consultation_mode, room_name, join_url, scheduled_at, duration_minutes,
		       status, payment_status, payment_due_at, paid_at, notes, created_at, updated_at
		FROM appointments `
	query += clause

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Appointment
	for rows.Next() {
		var a model.Appointment
		var paymentDueAt sql.NullTime
		var paidAt sql.NullTime
		var slotID sql.NullString
		var roomName sql.NullString
		var joinURL sql.NullString
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.PatientFirstName, &a.PatientLastName, &a.DoctorID, &a.DoctorOwnerUserID, &slotID, &a.ConsultationMode, &roomName, &joinURL, &a.ScheduledAt, &a.DurationMinutes,
			&a.Status, &a.PaymentStatus, &paymentDueAt, &paidAt, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if paymentDueAt.Valid {
			t := paymentDueAt.Time
			a.PaymentDueAt = &t
		}
		if paidAt.Valid {
			t := paidAt.Time
			a.PaidAt = &t
		}
		if slotID.Valid {
			a.SlotID = slotID.String
		}
		if roomName.Valid {
			a.RoomName = roomName.String
		}
		if joinURL.Valid {
			a.JoinURL = joinURL.String
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AppointmentRepository) UpdateStatus(id string, status model.AppointmentStatus) error {
	_, err := r.db.Exec(`UPDATE appointments SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *AppointmentRepository) MarkPaymentCompleted(id string) error {
	_, err := r.db.Exec(`
		UPDATE appointments
		SET payment_status = $1,
		    paid_at = NOW(),
		    status = CASE WHEN status = $2 THEN $2 ELSE $3 END,
		    updated_at = NOW()
		WHERE id = $4 AND status <> $5`,
		model.PaymentPaid, model.StatusCancelled, model.StatusConfirmed, id, model.StatusCancelled,
	)
	return err
}

func (r *AppointmentRepository) FindOverdueUnpaid(now time.Time) ([]model.Appointment, error) {
	rows, err := r.db.Query(`
		SELECT id, patient_id, patient_first_name, patient_last_name, doctor_id, doctor_owner_user_id, slot_id, consultation_mode, room_name, join_url, scheduled_at, duration_minutes,
		       status, payment_status, payment_due_at, paid_at, notes, created_at, updated_at
		FROM appointments
		WHERE payment_status = $1
		  AND payment_due_at IS NOT NULL
		  AND payment_due_at <= $2
		  AND status IN ($3, $4)`,
		model.PaymentPending, now, model.StatusPending, model.StatusConfirmed,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Appointment
	for rows.Next() {
		var a model.Appointment
		var paymentDueAt sql.NullTime
		var paidAt sql.NullTime
		var slotID sql.NullString
		var roomName sql.NullString
		var joinURL sql.NullString
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.PatientFirstName, &a.PatientLastName, &a.DoctorID, &a.DoctorOwnerUserID, &slotID, &a.ConsultationMode, &roomName, &joinURL, &a.ScheduledAt, &a.DurationMinutes,
			&a.Status, &a.PaymentStatus, &paymentDueAt, &paidAt, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if paymentDueAt.Valid {
			t := paymentDueAt.Time
			a.PaymentDueAt = &t
		}
		if paidAt.Valid {
			t := paidAt.Time
			a.PaidAt = &t
		}
		if slotID.Valid {
			a.SlotID = slotID.String
		}
		if roomName.Valid {
			a.RoomName = roomName.String
		}
		if joinURL.Valid {
			a.JoinURL = joinURL.String
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AppointmentRepository) CancelAndRelease(id string, paymentStatus model.PaymentStatus) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var slotID sql.NullString
	err = tx.QueryRow(`SELECT slot_id FROM appointments WHERE id = $1 FOR UPDATE`, id).Scan(&slotID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE appointments
		SET status = $1,
		    payment_status = $2,
		    updated_at = NOW()
		WHERE id = $3`,
		model.StatusCancelled, paymentStatus, id,
	); err != nil {
		return err
	}

	if slotID.Valid {
		if _, err := tx.Exec(`UPDATE slots SET is_booked = FALSE WHERE id = $1`, slotID.String); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AppointmentRepository) ReleaseSlot(slotID string) error {
	_, err := r.db.Exec(`UPDATE slots SET is_booked = FALSE WHERE id = $1`, slotID)
	return err
}

func (r *AppointmentRepository) GetSlotsByDoctor(doctorID, status string) ([]model.Slot, error) {
	query := `
		SELECT id, doctor_id, owner_user_id, start_time, end_time, is_booked, hospital
		FROM slots
		WHERE doctor_id = $1
		  AND start_time > NOW()`

	switch status {
	case "available":
		query += ` AND is_booked = FALSE`
	case "booked":
		query += ` AND is_booked = TRUE`
	case "all":
		// No additional filter.
	default:
		return nil, fmt.Errorf("invalid slot status filter")
	}

	query += ` ORDER BY start_time ASC`

	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var s model.Slot
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.OwnerUserID, &s.StartTime, &s.EndTime, &s.IsBooked, &s.Hospital); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (r *AppointmentRepository) CreateSlot(s *model.Slot) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	conflict, err := r.hasSlotConflict(s.DoctorID, s.StartTime, s.EndTime, "")
	if err != nil {
		return err
	}
	if conflict {
		return fmt.Errorf("slot overlaps an existing slot for this doctor")
	}

	_, err = r.db.Exec(`
		INSERT INTO slots (id, doctor_id, owner_user_id, start_time, end_time, is_booked, hospital)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		s.ID, s.DoctorID, s.OwnerUserID, s.StartTime, s.EndTime, s.IsBooked, s.Hospital,
	)
	return err
}

func (r *AppointmentRepository) GetSlotByID(id string) (*model.Slot, error) {
	s := &model.Slot{}
	err := r.db.QueryRow(`
		SELECT id, doctor_id, owner_user_id, start_time, end_time, is_booked, hospital
		FROM slots WHERE id = $1`, id).
		Scan(&s.ID, &s.DoctorID, &s.OwnerUserID, &s.StartTime, &s.EndTime, &s.IsBooked, &s.Hospital)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *AppointmentRepository) UpdateSlot(s *model.Slot) error {
	conflict, err := r.hasSlotConflict(s.DoctorID, s.StartTime, s.EndTime, s.ID)
	if err != nil {
		return err
	}
	if conflict {
		return fmt.Errorf("slot overlaps an existing slot for this doctor")
	}

	_, err = r.db.Exec(`
		UPDATE slots
		SET start_time = $1, end_time = $2, is_booked = $3, hospital = $4
		WHERE id = $5`,
		s.StartTime, s.EndTime, s.IsBooked, s.Hospital, s.ID,
	)
	return err
}

func (r *AppointmentRepository) hasSlotConflict(doctorID string, startTime, endTime time.Time, excludeSlotID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM slots
			WHERE doctor_id = $1
			  AND start_time < $3
			  AND end_time > $2`
	args := []interface{}{doctorID, startTime, endTime}

	if excludeSlotID != "" {
		query += ` AND id <> $4`
		args = append(args, excludeSlotID)
	}

	query += `
		)`

	var exists bool
	if err := r.db.QueryRow(query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *AppointmentRepository) DeleteSlot(id string) error {
	_, err := r.db.Exec(`DELETE FROM slots WHERE id = $1`, id)
	return err
}

func (r *AppointmentRepository) ListSlots() ([]model.Slot, error) {
	rows, err := r.db.Query(`SELECT id, doctor_id, owner_user_id, start_time, end_time, is_booked, hospital FROM slots ORDER BY start_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var s model.Slot
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.OwnerUserID, &s.StartTime, &s.EndTime, &s.IsBooked, &s.Hospital); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (r *AppointmentRepository) ListSlotsByOwner(ownerUserID string) ([]model.Slot, error) {
	rows, err := r.db.Query(`
		SELECT id, doctor_id, owner_user_id, start_time, end_time, is_booked, hospital
		FROM slots
		WHERE owner_user_id = $1
		ORDER BY start_time DESC`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var s model.Slot
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.OwnerUserID, &s.StartTime, &s.EndTime, &s.IsBooked, &s.Hospital); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}
