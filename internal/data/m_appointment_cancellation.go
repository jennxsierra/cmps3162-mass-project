package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
)

var ErrAppointmentAlreadyCancelled = errors.New("appointment already cancelled")
var ErrAppointmentCancellationNotFound = errors.New("appointment cancellation not found")
var ErrStaffNotFound = errors.New("staff not found")
var ErrCancellationReasonNotFound = errors.New("cancellation reason not found")

type AppointmentCancellation struct {
	AppointmentID int64     `json:"appointment_id"`
	CancelledAt   time.Time `json:"cancelled_at"`
	Note          string    `json:"note,omitempty"`
	ReasonID      int       `json:"reason_id"`
	ReasonName    string    `json:"reason_name,omitempty"`
	RecordedBy    int64     `json:"recorded_by"`
}

type AppointmentCancellationModel struct {
	DB *sql.DB
}

func ValidateAppointmentCancellation(v *validator.Validator, c *AppointmentCancellation) {
	v.Check(c.ReasonID > 0, "reason_id", "must be provided")
	v.Check(c.RecordedBy > 0, "recorded_by", "must be provided")
	v.Check(len(c.Note) <= 500, "note", "must not be more than 500 characters long")
}

func (m AppointmentCancellationModel) ensureAppointmentExists(ctx context.Context, appointmentID int64) error {
	var id int64
	err := m.DB.QueryRowContext(ctx, `SELECT appointment_id FROM appointment WHERE appointment_id = $1`, appointmentID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentCancellationModel) ensureCancellationReasonExists(ctx context.Context, reasonID int) error {
	var id int
	err := m.DB.QueryRowContext(ctx, `SELECT reason_id FROM cancellation_reason WHERE reason_id = $1`, reasonID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCancellationReasonNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentCancellationModel) ensureStaffExists(ctx context.Context, staffID int64) error {
	var id int64
	err := m.DB.QueryRowContext(ctx, `SELECT staff_id FROM staff WHERE staff_id = $1`, staffID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrStaffNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentCancellationModel) Insert(cancellation *AppointmentCancellation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.ensureAppointmentExists(ctx, cancellation.AppointmentID); err != nil {
		return err
	}
	if err := m.ensureCancellationReasonExists(ctx, cancellation.ReasonID); err != nil {
		return err
	}
	if err := m.ensureStaffExists(ctx, cancellation.RecordedBy); err != nil {
		return err
	}

	alreadyCancelledQuery := `
		SELECT EXISTS(
			SELECT 1 FROM appt_cancellation WHERE appointment_id = $1
		)
	`
	var alreadyCancelled bool
	err := m.DB.QueryRowContext(ctx, alreadyCancelledQuery, cancellation.AppointmentID).Scan(&alreadyCancelled)
	if err != nil {
		return err
	}
	if alreadyCancelled {
		return ErrAppointmentAlreadyCancelled
	}

	query := `
		INSERT INTO appt_cancellation (appointment_id, cancelled_at, note, reason_id, recorded_by)
		VALUES ($1, COALESCE($2, NOW()), NULLIF($3, ''), $4, $5)
		RETURNING appointment_id, cancelled_at, COALESCE(note, ''), reason_id, recorded_by
	`

	err = m.DB.QueryRowContext(
		ctx,
		query,
		cancellation.AppointmentID,
		nullableTime(cancellation.CancelledAt),
		strings.TrimSpace(cancellation.Note),
		cancellation.ReasonID,
		cancellation.RecordedBy,
	).Scan(
		&cancellation.AppointmentID,
		&cancellation.CancelledAt,
		&cancellation.Note,
		&cancellation.ReasonID,
		&cancellation.RecordedBy,
	)
	if err != nil {
		return err
	}

	return nil
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func (m AppointmentCancellationModel) GetByAppointment(appointmentID int64) (*AppointmentCancellation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.ensureAppointmentExists(ctx, appointmentID); err != nil {
		return nil, err
	}

	query := `
		SELECT
			ac.appointment_id,
			ac.cancelled_at,
			COALESCE(ac.note, ''),
			ac.reason_id,
			cr.reason_name,
			ac.recorded_by
		FROM appt_cancellation ac
		JOIN cancellation_reason cr ON cr.reason_id = ac.reason_id
		WHERE ac.appointment_id = $1
	`

	var cancellation AppointmentCancellation
	err := m.DB.QueryRowContext(ctx, query, appointmentID).Scan(
		&cancellation.AppointmentID,
		&cancellation.CancelledAt,
		&cancellation.Note,
		&cancellation.ReasonID,
		&cancellation.ReasonName,
		&cancellation.RecordedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppointmentCancellationNotFound
		}
		return nil, err
	}

	return &cancellation, nil
}
