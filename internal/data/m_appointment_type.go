package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
)

type AppointmentType struct {
	AppointmentTypeID int    `json:"appointment_type_id"`
	AppointmentName   string `json:"appointment_name"`
}

type AppointmentTypeModel struct {
	DB *sql.DB
}

func (m AppointmentTypeModel) Insert(appointmentType *AppointmentType) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO appt_type (appt_type_name)
		VALUES ($1)
		RETURNING appt_type_id
	`

	err := m.DB.QueryRowContext(ctx, query, appointmentType.AppointmentName).Scan(&appointmentType.AppointmentTypeID)
	if err != nil {
		return err
	}

	return nil
}

func (m AppointmentTypeModel) Get(id int) (*AppointmentType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT appt_type_id, appt_type_name
		FROM appt_type
		WHERE appt_type_id = $1
	`

	var at AppointmentType

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&at.AppointmentTypeID, &at.AppointmentName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &at, nil
}

func (m AppointmentTypeModel) GetAll(filters Filters) ([]*AppointmentType, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), appt_type_id, appt_type_name
		FROM appt_type
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	appointmentTypes := []*AppointmentType{}

	for rows.Next() {
		var at AppointmentType

		err := rows.Scan(&totalRecords, &at.AppointmentTypeID, &at.AppointmentName)
		if err != nil {
			return nil, Metadata{}, err
		}

		appointmentTypes = append(appointmentTypes, &at)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return appointmentTypes, metadata, nil
}

func (m AppointmentTypeModel) Update(appointmentType *AppointmentType) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE appt_type
		SET appt_type_name = $1
		WHERE appt_type_id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, appointmentType.AppointmentName, appointmentType.AppointmentTypeID)
	if err != nil {
		return err
	}

	return nil
}

func (m AppointmentTypeModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM appt_type
		WHERE appt_type_id = $1
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// ValidateAppointmentType validation
func ValidateAppointmentType(v *validator.Validator, at *AppointmentType) {
	v.Check(at.AppointmentName != "", "appointment_name", "must be provided")
	v.Check(len(at.AppointmentName) <= 100, "appointment_name", "must not be more than 100 characters long")
}
