package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
)

type Specialty struct {
	SpecialtyID   int    `json:"specialty_id"`
	SpecialtyName string `json:"specialty_name"`
}

type SpecialtyModel struct {
	DB *sql.DB
}

func (m SpecialtyModel) Insert(specialty *Specialty) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO specialty (specialty_name)
		VALUES ($1)
		RETURNING specialty_id
	`

	err := m.DB.QueryRowContext(ctx, query, specialty.SpecialtyName).Scan(&specialty.SpecialtyID)
	if err != nil {
		return err
	}

	return nil
}

func (m SpecialtyModel) Get(id int) (*Specialty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT specialty_id, specialty_name
		FROM specialty
		WHERE specialty_id = $1
	`

	var s Specialty

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&s.SpecialtyID, &s.SpecialtyName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &s, nil
}

func (m SpecialtyModel) GetAll(filters Filters) ([]*Specialty, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), specialty_id, specialty_name
		FROM specialty
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	specialties := []*Specialty{}

	for rows.Next() {
		var s Specialty

		err := rows.Scan(&totalRecords, &s.SpecialtyID, &s.SpecialtyName)
		if err != nil {
			return nil, Metadata{}, err
		}

		specialties = append(specialties, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return specialties, metadata, nil
}

func (m SpecialtyModel) Update(specialty *Specialty) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE specialty
		SET specialty_name = $1
		WHERE specialty_id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, specialty.SpecialtyName, specialty.SpecialtyID)
	if err != nil {
		return err
	}

	return nil
}

func (m SpecialtyModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM specialty
		WHERE specialty_id = $1
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// ValidateSpecialty validation
func ValidateSpecialty(v *validator.Validator, s *Specialty) {
	v.Check(s.SpecialtyName != "", "specialty_name", "must be provided")
	v.Check(len(s.SpecialtyName) <= 100, "specialty_name", "must not be more than 100 characters long")
}
