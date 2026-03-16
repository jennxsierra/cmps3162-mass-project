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

var ErrProviderInUse = errors.New("provider has related records and cannot be deleted")

type Provider struct {
	ProviderID  int64     `json:"provider_id"`
	LicenseNo   string    `json:"license_no"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	CreatedAt   time.Time `json:"created_at"`
}

func ValidateProvider(v *validator.Validator, p *Provider) {
	v.Check(p.LicenseNo != "", "license_no", "License number must be provided")
	v.Check(p.FirstName != "", "first_name", "First name must be provided")
	v.Check(p.LastName != "", "last_name", "Last name must be provided")

	if p.DateOfBirth != "" {
		_, err := time.Parse("2006-01-02", p.DateOfBirth)
		v.Check(err == nil, "date_of_birth", "Date of birth must be a valid date (YYYY-MM-DD)")
	}
}

type ProviderModel struct {
	DB *sql.DB
}

func (m ProviderModel) Insert(p *Provider) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	personQuery := `
		INSERT INTO person (first_name, last_name, date_of_birth, gender, created_at)
		VALUES ($1, $2, NULLIF($3, '')::date, NULLIF($4, ''), NOW())
		RETURNING person_id, created_at
	`
	err := m.DB.QueryRowContext(ctx, personQuery, p.FirstName, p.LastName, p.DateOfBirth, p.Gender).Scan(&p.ProviderID, &p.CreatedAt)
	if err != nil {
		return err
	}

	providerQuery := `
		INSERT INTO provider (provider_id, license_no)
		VALUES ($1, $2)
	`
	_, err = m.DB.ExecContext(ctx, providerQuery, p.ProviderID, p.LicenseNo)
	return err
}

func (m ProviderModel) Get(licenseNo string) (*Provider, error) {
	query := `
		SELECT
			pr.provider_id, pr.license_no,
			pe.first_name, pe.last_name, pe.date_of_birth, pe.gender, pe.created_at
		FROM provider pr
		JOIN person pe ON pr.provider_id = pe.person_id
		WHERE pr.license_no = $1
	`

	var p Provider
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, licenseNo).Scan(
		&p.ProviderID, &p.LicenseNo,
		&p.FirstName, &p.LastName, &p.DateOfBirth, &p.Gender, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &p, nil
}

func (m ProviderModel) GetAll(search string, specialtyID int, filters Filters) ([]*Provider, Metadata, error) {
	query := fmt.Sprintf(`
		WITH filtered AS (
			SELECT DISTINCT
				pr.provider_id,
				pr.license_no,
				pe.first_name,
				pe.last_name,
				pe.date_of_birth,
				pe.gender,
				pe.created_at
			FROM provider pr
			JOIN person pe ON pr.provider_id = pe.person_id
			LEFT JOIN provider_specialty ps ON pr.provider_id = ps.provider_id
			WHERE ($1 = ''
				OR pe.first_name ILIKE '%%' || $1 || '%%'
				OR pe.last_name ILIKE '%%' || $1 || '%%'
				OR (pe.first_name || ' ' || pe.last_name) ILIKE '%%' || $1 || '%%'
				OR pr.license_no ILIKE '%%' || $1 || '%%')
			AND ($2 = 0 OR ps.specialty_id = $2)
		)
		SELECT
			COUNT(*) OVER(),
			f.provider_id,
			f.license_no,
			f.first_name,
			f.last_name,
			f.date_of_birth,
			f.gender,
			f.created_at
		FROM filtered f
		ORDER BY %s %s, f.provider_id ASC
		LIMIT $3 OFFSET $4
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, search, specialtyID, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	providers := []*Provider{}

	for rows.Next() {
		var p Provider
		err := rows.Scan(
			&totalRecords,
			&p.ProviderID,
			&p.LicenseNo,
			&p.FirstName,
			&p.LastName,
			&p.DateOfBirth,
			&p.Gender,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		providers = append(providers, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return providers, metadata, nil
}

func (m ProviderModel) Update(p *Provider) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	personQuery := `
		UPDATE person
		SET first_name = $1, last_name = $2, date_of_birth = NULLIF($3, '')::date, gender = NULLIF($4, '')
		WHERE person_id = $5
	`
	_, err := m.DB.ExecContext(ctx, personQuery, p.FirstName, p.LastName, p.DateOfBirth, p.Gender, p.ProviderID)
	if err != nil {
		return err
	}

	providerQuery := `
		UPDATE provider
		SET license_no = $1
		WHERE provider_id = $2
	`
	result, err := m.DB.ExecContext(ctx, providerQuery, p.LicenseNo, p.ProviderID)
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

func (m ProviderModel) Delete(licenseNo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx,
		"DELETE FROM person WHERE person_id = (SELECT provider_id FROM provider WHERE license_no = $1)",
		licenseNo,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return ErrProviderInUse
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
