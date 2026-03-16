package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jennxsierra/rate-limiting-shutdown/internal/validator"
)

type Patient struct {
	PatientID   int64     `json:"patient_id"`
	PatientNo   string    `json:"patient_no"`
	SSN         string    `json:"ssn"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	CreatedAt   time.Time `json:"created_at"`
}

func ValidatePatient(v *validator.Validator, p *Patient) {
	v.Check(p.PatientNo != "", "patient_no", "Patient number must be provided")
	v.Check(p.FirstName != "", "first_name", "First name must be provided")
	v.Check(p.LastName != "", "last_name", "Last name must be provided")
	v.Check(p.SSN != "", "ssn", "SSN must be provided")
}

type PatientModel struct {
	DB *sql.DB
}

// Insert a patient and linked person record
func (m PatientModel) Insert(p *Patient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Insert into person
	personQuery := `
        INSERT INTO person (first_name, last_name, date_of_birth, gender, created_at)
        VALUES ($1, $2, $3, $4, NOW())
        RETURNING person_id, created_at
    `
	err := m.DB.QueryRowContext(ctx, personQuery, p.FirstName, p.LastName, p.DateOfBirth, p.Gender).Scan(&p.PatientID, &p.CreatedAt)
	if err != nil {
		return err
	}

	// Insert into patient
	patientQuery := `
        INSERT INTO patient (patient_id, patient_no, ssn)
        VALUES ($1, $2, $3)
    `
	_, err = m.DB.ExecContext(ctx, patientQuery, p.PatientID, p.PatientNo, p.SSN)
	return err
}

// Get a patient by patient_no (JOIN person and patient)
func (m PatientModel) Get(patientNo string) (*Patient, error) {
	query := `
        SELECT 
            pa.patient_id, pa.patient_no, pa.ssn, 
            pe.first_name, pe.last_name, pe.date_of_birth, pe.gender, pe.created_at
        FROM patient pa
        JOIN person pe ON pa.patient_id = pe.person_id
        WHERE pa.patient_no = $1
    `
	var p Patient
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, patientNo).Scan(
		&p.PatientID, &p.PatientNo, &p.SSN,
		&p.FirstName, &p.LastName, &p.DateOfBirth, &p.Gender, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("record not found")
		}
		return nil, err
	}
	return &p, nil
}

// Get specific patients based on the query parameters (first_name, last_name, and pagination)
func (m PatientModel) GetAll(firstName string, lastName string, filters Filters) ([]*Patient, Metadata, error) {
	// $? = '' allows for firstName and lastName to be optional
	// $3 and $4 are LIMIT and OFFSET for pagination
	query := fmt.Sprintf(`
        SELECT 
			COUNT(*) OVER(),
			pa.patient_id, pa.patient_no, pa.ssn, 
            pe.first_name, pe.last_name, pe.date_of_birth, pe.gender, pe.created_at
        FROM patient pa
        JOIN person pe ON pa.patient_id = pe.person_id
        WHERE (to_tsvector('simple', pe.first_name) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
        AND (to_tsvector('simple', pe.last_name) @@ 
             plainto_tsquery('simple', $2) OR $2 = '') 
        ORDER BY %s %s, pa.patient_id ASC
        LIMIT $3 OFFSET $4
    `, filters.sortColumn(), filters.sortDirection())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, firstName, lastName, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var results []*Patient
	for rows.Next() {
		var p Patient
		err := rows.Scan(
			&totalRecords,
			&p.PatientID, &p.PatientNo, &p.SSN,
			&p.FirstName, &p.LastName, &p.DateOfBirth, &p.Gender, &p.CreatedAt)
		if err != nil {
			return nil, Metadata{}, err
		}
		results = append(results, &p)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return results, metadata, nil
}

// Update patient info (PATCH semantics: update only provided non-nil fields)
func (m PatientModel) Update(p *Patient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Update person details
	personQuery := `
        UPDATE person SET first_name=$1, last_name=$2, date_of_birth=$3, gender=$4
        WHERE person_id=$5
    `
	_, err := m.DB.ExecContext(ctx, personQuery, p.FirstName, p.LastName, p.DateOfBirth, p.Gender, p.PatientID)
	if err != nil {
		return err
	}

	// Update patient details
	patientQuery := `
        UPDATE patient SET ssn=$1
        WHERE patient_id=$2
    `
	result, err := m.DB.ExecContext(ctx, patientQuery, p.SSN, p.PatientID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found")
	}
	return nil
}

// Delete patient and linked person (cascade delete)
func (m PatientModel) Delete(patientNo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Find person_id to delete
	var personID int64
	err := m.DB.QueryRowContext(ctx, "SELECT patient_id FROM patient WHERE patient_no=$1", patientNo).Scan(&personID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("record not found")
		}
		return err
	}

	// Delete from person, which cascades to patient
	result, err := m.DB.ExecContext(ctx, "DELETE FROM person WHERE person_id=$1", personID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found")
	}
	return nil
}
