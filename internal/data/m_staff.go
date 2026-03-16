package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
	"github.com/lib/pq"
)

var ErrStaffInUse = errors.New("staff has related records and cannot be deleted")

type Staff struct {
	StaffID     int64     `json:"staff_id"`
	StaffNo     string    `json:"staff_no"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	CreatedAt   time.Time `json:"created_at"`
}

func ValidateStaff(v *validator.Validator, staff *Staff) {
	v.Check(staff.StaffNo != "", "staff_no", "Staff number must be provided")
	v.Check(staff.FirstName != "", "first_name", "First name must be provided")
	v.Check(staff.LastName != "", "last_name", "Last name must be provided")

	if staff.DateOfBirth != "" {
		_, err := time.Parse("2006-01-02", staff.DateOfBirth)
		v.Check(err == nil, "date_of_birth", "Date of birth must be a valid date (YYYY-MM-DD)")
	}
}

type StaffModel struct {
	DB *sql.DB
}

func (m StaffModel) Insert(staff *Staff) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	personQuery := `
		INSERT INTO person (first_name, last_name, date_of_birth, gender, created_at)
		VALUES ($1, $2, NULLIF($3, '')::date, NULLIF($4, ''), NOW())
		RETURNING person_id, created_at
	`
	err := m.DB.QueryRowContext(ctx, personQuery,
		staff.FirstName, staff.LastName, staff.DateOfBirth, staff.Gender,
	).Scan(&staff.StaffID, &staff.CreatedAt)
	if err != nil {
		return err
	}

	staffQuery := `
		INSERT INTO staff (staff_id, staff_no)
		VALUES ($1, $2)
	`
	_, err = m.DB.ExecContext(ctx, staffQuery, staff.StaffID, staff.StaffNo)
	return err
}

func (m StaffModel) Get(staffNo string) (*Staff, error) {
	query := `
		SELECT
			s.staff_id, s.staff_no,
			p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
		FROM staff s
		JOIN person p ON p.person_id = s.staff_id
		WHERE s.staff_no = $1
	`

	var staff Staff
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, staffNo).Scan(
		&staff.StaffID, &staff.StaffNo,
		&staff.FirstName, &staff.LastName, &staff.DateOfBirth, &staff.Gender, &staff.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &staff, nil
}

func (m StaffModel) GetAll(search string, filters Filters) ([]*Staff, Metadata, error) {
	query := fmt.Sprintf(`
		WITH filtered AS (
			SELECT
				s.staff_id,
				s.staff_no,
				p.first_name,
				p.last_name,
				p.date_of_birth,
				p.gender,
				p.created_at
			FROM staff s
			JOIN person p ON p.person_id = s.staff_id
			WHERE ($1 = ''
				OR p.first_name ILIKE '%%' || $1 || '%%'
				OR p.last_name ILIKE '%%' || $1 || '%%'
				OR (p.first_name || ' ' || p.last_name) ILIKE '%%' || $1 || '%%'
				OR s.staff_no ILIKE '%%' || $1 || '%%')
		)
		SELECT
			COUNT(*) OVER(),
			f.staff_id,
			f.staff_no,
			f.first_name,
			f.last_name,
			f.date_of_birth,
			f.gender,
			f.created_at
		FROM filtered f
		ORDER BY %s %s, f.staff_id ASC
		LIMIT $2 OFFSET $3
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, search, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	staffMembers := []*Staff{}

	for rows.Next() {
		var staff Staff
		err := rows.Scan(
			&totalRecords,
			&staff.StaffID,
			&staff.StaffNo,
			&staff.FirstName,
			&staff.LastName,
			&staff.DateOfBirth,
			&staff.Gender,
			&staff.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		staffMembers = append(staffMembers, &staff)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return staffMembers, metadata, nil
}

func (m StaffModel) Update(staff *Staff) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	personQuery := `
		UPDATE person
		SET first_name = $1, last_name = $2, date_of_birth = NULLIF($3, '')::date, gender = NULLIF($4, '')
		WHERE person_id = $5
	`
	_, err := m.DB.ExecContext(ctx, personQuery,
		staff.FirstName, staff.LastName, staff.DateOfBirth, staff.Gender, staff.StaffID,
	)
	if err != nil {
		return err
	}

	staffQuery := `
		UPDATE staff
		SET staff_no = $1
		WHERE staff_id = $2
	`
	result, err := m.DB.ExecContext(ctx, staffQuery, staff.StaffNo, staff.StaffID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m StaffModel) Delete(staffNo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx,
		"DELETE FROM person WHERE person_id = (SELECT staff_id FROM staff WHERE staff_no = $1)",
		staffNo,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return ErrStaffInUse
		}
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}

	return nil
}
