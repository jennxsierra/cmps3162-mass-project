package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
)

var ErrInvalidAppointmentTimeRange = errors.New("end_time must be greater than start_time")
var ErrProviderUnavailable = errors.New("provider has overlapping appointment")
var ErrPatientUnavailable = errors.New("patient has overlapping appointment")
var ErrCancelledAppointmentScheduleChange = errors.New("cannot change schedule fields for cancelled appointment")

type Appointment struct {
	AppointmentID int64      `json:"appointment_id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	PatientID     int64      `json:"patient_id"`
	ProviderID    int64      `json:"provider_id"`
	CreatedBy     int64      `json:"created_by"`
	ApptTypeID    int        `json:"appt_type_id"`
	ApptTypeName  string     `json:"appt_type_name,omitempty"`
	PatientName   string     `json:"patient_name,omitempty"`
	ProviderName  string     `json:"provider_name,omitempty"`
	IsCancelled   bool       `json:"is_cancelled"`
}

type AppointmentQueryParams struct {
	ProviderID       int64
	PatientID        int64
	ApptTypeID       int
	StartFrom        *time.Time
	StartTo          *time.Time
	IncludeCancelled bool
}

type AppointmentModel struct {
	DB *sql.DB
}

func ValidateAppointment(v *validator.Validator, a *Appointment) {
	v.Check(!a.StartTime.IsZero(), "start_time", "must be provided")
	v.Check(!a.EndTime.IsZero(), "end_time", "must be provided")
	if !a.StartTime.IsZero() && !a.EndTime.IsZero() {
		v.Check(a.EndTime.After(a.StartTime), "end_time", "must be greater than start_time")
	}
	v.Check(a.PatientID > 0, "patient_id", "must be provided")
	v.Check(a.ProviderID > 0, "provider_id", "must be provided")
	v.Check(a.CreatedBy > 0, "created_by", "must be provided")
	v.Check(a.ApptTypeID > 0, "appt_type_id", "must be provided")
	v.Check(len(a.Reason) <= 500, "reason", "must not be more than 500 characters long")
}

func (m AppointmentModel) ensurePatientExists(ctx context.Context, id int64) error {
	var v int64
	err := m.DB.QueryRowContext(ctx, `SELECT patient_id FROM patient WHERE patient_id = $1`, id).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentModel) ensureProviderExists(ctx context.Context, id int64) error {
	var v int64
	err := m.DB.QueryRowContext(ctx, `SELECT provider_id FROM provider WHERE provider_id = $1`, id).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentModel) ensureStaffExists(ctx context.Context, id int64) error {
	var v int64
	err := m.DB.QueryRowContext(ctx, `SELECT staff_id FROM staff WHERE staff_id = $1`, id).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentModel) ensureApptTypeExists(ctx context.Context, id int) error {
	var v int
	err := m.DB.QueryRowContext(ctx, `SELECT appt_type_id FROM appt_type WHERE appt_type_id = $1`, id).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m AppointmentModel) providerHasOverlap(ctx context.Context, providerID int64, start, end time.Time, ignoreID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM appointment a
			LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
			WHERE a.provider_id = $1
			  AND ac.appointment_id IS NULL
			  AND ($2 = 0 OR a.appointment_id <> $2)
			  AND a.start_time < $3
			  AND a.end_time > $4
		)
	`
	var hasOverlap bool
	err := m.DB.QueryRowContext(ctx, query, providerID, ignoreID, end, start).Scan(&hasOverlap)
	return hasOverlap, err
}

func (m AppointmentModel) patientHasOverlap(ctx context.Context, patientID int64, start, end time.Time, ignoreID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM appointment a
			LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
			WHERE a.patient_id = $1
			  AND ac.appointment_id IS NULL
			  AND ($2 = 0 OR a.appointment_id <> $2)
			  AND a.start_time < $3
			  AND a.end_time > $4
		)
	`
	var hasOverlap bool
	err := m.DB.QueryRowContext(ctx, query, patientID, ignoreID, end, start).Scan(&hasOverlap)
	return hasOverlap, err
}

func (m AppointmentModel) IsCancelled(id int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT EXISTS (SELECT 1 FROM appt_cancellation WHERE appointment_id = $1)`
	var cancelled bool
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&cancelled)
	if err != nil {
		return false, err
	}
	return cancelled, nil
}

func (m AppointmentModel) Insert(a *Appointment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if !a.EndTime.After(a.StartTime) {
		return ErrInvalidAppointmentTimeRange
	}

	if err := m.ensurePatientExists(ctx, a.PatientID); err != nil {
		return err
	}
	if err := m.ensureProviderExists(ctx, a.ProviderID); err != nil {
		return err
	}
	if err := m.ensureStaffExists(ctx, a.CreatedBy); err != nil {
		return err
	}
	if err := m.ensureApptTypeExists(ctx, a.ApptTypeID); err != nil {
		return err
	}

	hasProviderOverlap, err := m.providerHasOverlap(ctx, a.ProviderID, a.StartTime, a.EndTime, 0)
	if err != nil {
		return err
	}
	if hasProviderOverlap {
		return ErrProviderUnavailable
	}

	hasPatientOverlap, err := m.patientHasOverlap(ctx, a.PatientID, a.StartTime, a.EndTime, 0)
	if err != nil {
		return err
	}
	if hasPatientOverlap {
		return ErrPatientUnavailable
	}

	query := `
		INSERT INTO appointment (
			start_time,
			end_time,
			reason,
			patient_id,
			provider_id,
			created_by,
			appt_type_id
		)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7)
		RETURNING appointment_id, created_at
	`

	err = m.DB.QueryRowContext(
		ctx,
		query,
		a.StartTime,
		a.EndTime,
		a.Reason,
		a.PatientID,
		a.ProviderID,
		a.CreatedBy,
		a.ApptTypeID,
	).Scan(&a.AppointmentID, &a.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (m AppointmentModel) Get(id int64) (*Appointment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT
			a.appointment_id,
			a.start_time,
			a.end_time,
			a.created_at,
			a.updated_at,
			COALESCE(a.reason, ''),
			a.patient_id,
			a.provider_id,
			a.created_by,
			a.appt_type_id,
			atp.appt_type_name,
			(pp.first_name || ' ' || pp.last_name) AS patient_name,
			(prp.first_name || ' ' || prp.last_name) AS provider_name,
			(ac.appointment_id IS NOT NULL) AS is_cancelled
		FROM appointment a
		JOIN appt_type atp ON atp.appt_type_id = a.appt_type_id
		JOIN person pp ON pp.person_id = a.patient_id
		JOIN person prp ON prp.person_id = a.provider_id
		LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
		WHERE a.appointment_id = $1
	`

	var appointment Appointment
	var updatedAt sql.NullTime

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&appointment.AppointmentID,
		&appointment.StartTime,
		&appointment.EndTime,
		&appointment.CreatedAt,
		&updatedAt,
		&appointment.Reason,
		&appointment.PatientID,
		&appointment.ProviderID,
		&appointment.CreatedBy,
		&appointment.ApptTypeID,
		&appointment.ApptTypeName,
		&appointment.PatientName,
		&appointment.ProviderName,
		&appointment.IsCancelled,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if updatedAt.Valid {
		appointment.UpdatedAt = &updatedAt.Time
	}

	return &appointment, nil
}

func (m AppointmentModel) GetAll(queryParams AppointmentQueryParams, filters Filters) ([]*Appointment, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) OVER(),
			a.appointment_id,
			a.start_time,
			a.end_time,
			a.created_at,
			a.updated_at,
			COALESCE(a.reason, ''),
			a.patient_id,
			a.provider_id,
			a.created_by,
			a.appt_type_id,
			(ac.appointment_id IS NOT NULL) AS is_cancelled
		FROM appointment a
		LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
		WHERE ($1 = 0 OR a.provider_id = $1)
		  AND ($2 = 0 OR a.patient_id = $2)
		  AND ($3 = 0 OR a.appt_type_id = $3)
		  AND ($4::timestamptz IS NULL OR a.start_time >= $4)
		  AND ($5::timestamptz IS NULL OR a.start_time <= $5)
		  AND ($6 OR ac.appointment_id IS NULL)
		ORDER BY %s %s, a.appointment_id ASC
		LIMIT $7 OFFSET $8
	`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(
		ctx,
		query,
		queryParams.ProviderID,
		queryParams.PatientID,
		queryParams.ApptTypeID,
		queryParams.StartFrom,
		queryParams.StartTo,
		queryParams.IncludeCancelled,
		filters.limit(),
		filters.offset(),
	)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	appointments := []*Appointment{}

	for rows.Next() {
		var appointment Appointment
		var updatedAt sql.NullTime

		err := rows.Scan(
			&totalRecords,
			&appointment.AppointmentID,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.CreatedAt,
			&updatedAt,
			&appointment.Reason,
			&appointment.PatientID,
			&appointment.ProviderID,
			&appointment.CreatedBy,
			&appointment.ApptTypeID,
			&appointment.IsCancelled,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		if updatedAt.Valid {
			appointment.UpdatedAt = &updatedAt.Time
		}

		appointments = append(appointments, &appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return appointments, metadata, nil
}

func (m AppointmentModel) Update(a *Appointment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	existing, err := m.Get(a.AppointmentID)
	if err != nil {
		return err
	}

	if existing.IsCancelled {
		scheduleChanged :=
			!a.StartTime.Equal(existing.StartTime) ||
				!a.EndTime.Equal(existing.EndTime) ||
				a.PatientID != existing.PatientID ||
				a.ProviderID != existing.ProviderID ||
				a.ApptTypeID != existing.ApptTypeID
		if scheduleChanged {
			return ErrCancelledAppointmentScheduleChange
		}
	}

	if !a.EndTime.After(a.StartTime) {
		return ErrInvalidAppointmentTimeRange
	}

	if err := m.ensurePatientExists(ctx, a.PatientID); err != nil {
		return err
	}
	if err := m.ensureProviderExists(ctx, a.ProviderID); err != nil {
		return err
	}
	if err := m.ensureStaffExists(ctx, a.CreatedBy); err != nil {
		return err
	}
	if err := m.ensureApptTypeExists(ctx, a.ApptTypeID); err != nil {
		return err
	}

	hasProviderOverlap, err := m.providerHasOverlap(ctx, a.ProviderID, a.StartTime, a.EndTime, a.AppointmentID)
	if err != nil {
		return err
	}
	if hasProviderOverlap {
		return ErrProviderUnavailable
	}

	hasPatientOverlap, err := m.patientHasOverlap(ctx, a.PatientID, a.StartTime, a.EndTime, a.AppointmentID)
	if err != nil {
		return err
	}
	if hasPatientOverlap {
		return ErrPatientUnavailable
	}

	query := `
		UPDATE appointment
		SET start_time = $1,
			end_time = $2,
			reason = NULLIF($3, ''),
			patient_id = $4,
			provider_id = $5,
			created_by = $6,
			appt_type_id = $7,
			updated_at = NOW()
		WHERE appointment_id = $8
	`

	result, err := m.DB.ExecContext(
		ctx,
		query,
		a.StartTime,
		a.EndTime,
		a.Reason,
		a.PatientID,
		a.ProviderID,
		a.CreatedBy,
		a.ApptTypeID,
		a.AppointmentID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
