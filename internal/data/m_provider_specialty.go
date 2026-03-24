package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrProviderNotFound = errors.New("provider not found")
var ErrSpecialtyNotFound = errors.New("specialty not found")
var ErrProviderSpecialtyNotFound = errors.New("provider specialty mapping not found")

type ProviderSpecialty struct {
	ProviderID  int64 `json:"provider_id"`
	SpecialtyID int   `json:"specialty_id"`
}

type ProviderSpecialtyModel struct {
	DB *sql.DB
}

func (m ProviderSpecialtyModel) getProviderIDByLicense(ctx context.Context, licenseNo string) (int64, error) {
	query := `
		SELECT provider_id
		FROM provider
		WHERE license_no = $1
	`

	var providerID int64
	err := m.DB.QueryRowContext(ctx, query, licenseNo).Scan(&providerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrProviderNotFound
		}
		return 0, err
	}

	return providerID, nil
}

func (m ProviderSpecialtyModel) ensureSpecialtyExists(ctx context.Context, specialtyID int) error {
	query := `
		SELECT specialty_id
		FROM specialty
		WHERE specialty_id = $1
	`

	var id int
	err := m.DB.QueryRowContext(ctx, query, specialtyID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSpecialtyNotFound
		}
		return err
	}

	return nil
}

func (m ProviderSpecialtyModel) Insert(licenseNo string, specialtyID int) (*ProviderSpecialty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	providerID, err := m.getProviderIDByLicense(ctx, licenseNo)
	if err != nil {
		return nil, err
	}

	err = m.ensureSpecialtyExists(ctx, specialtyID)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO provider_specialty (provider_id, specialty_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err = m.DB.ExecContext(ctx, query, providerID, specialtyID)
	if err != nil {
		return nil, err
	}

	providerSpecialty := &ProviderSpecialty{
		ProviderID:  providerID,
		SpecialtyID: specialtyID,
	}

	return providerSpecialty, nil
}

func (m ProviderSpecialtyModel) GetAllByProvider(licenseNo string) ([]*Specialty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	providerID, err := m.getProviderIDByLicense(ctx, licenseNo)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT s.specialty_id, s.specialty_name
		FROM provider_specialty ps
		JOIN specialty s ON s.specialty_id = ps.specialty_id
		WHERE ps.provider_id = $1
		ORDER BY s.specialty_name
	`

	rows, err := m.DB.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	specialties := []*Specialty{}
	for rows.Next() {
		var specialty Specialty

		err := rows.Scan(&specialty.SpecialtyID, &specialty.SpecialtyName)
		if err != nil {
			return nil, err
		}

		specialties = append(specialties, &specialty)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return specialties, nil
}

func (m ProviderSpecialtyModel) Delete(licenseNo string, specialtyID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	providerID, err := m.getProviderIDByLicense(ctx, licenseNo)
	if err != nil {
		return err
	}

	query := `
		DELETE FROM provider_specialty
		WHERE provider_id = $1 AND specialty_id = $2
	`

	result, err := m.DB.ExecContext(ctx, query, providerID, specialtyID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProviderSpecialtyNotFound
	}

	return nil
}
