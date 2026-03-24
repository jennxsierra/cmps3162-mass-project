package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
)

var ErrPersonNotFound = errors.New("person not found")
var ErrContactTypeNotFound = errors.New("contact type not found")

type PersonContact struct {
	PersonContactID int64  `json:"person_contact_id"`
	PersonID        int64  `json:"person_id"`
	ContactTypeID   int    `json:"contact_type_id"`
	ContactTypeName string `json:"contact_type_name,omitempty"`
	ContactValue    string `json:"contact_value"`
	IsPrimary       bool   `json:"is_primary"`
}

type PersonContactModel struct {
	DB *sql.DB
}

func ValidatePersonContact(v *validator.Validator, contact *PersonContact) {
	v.Check(contact.PersonID > 0, "person_id", "must be provided")
	v.Check(contact.ContactTypeID > 0, "contact_type_id", "must be provided")
	v.Check(strings.TrimSpace(contact.ContactValue) != "", "contact_value", "must be provided")
	v.Check(len(strings.TrimSpace(contact.ContactValue)) <= 255, "contact_value", "must not be more than 255 characters long")
}

func (m PersonContactModel) ensurePersonExists(ctx context.Context, personID int64) error {
	query := `
		SELECT person_id
		FROM person
		WHERE person_id = $1
	`

	var id int64
	err := m.DB.QueryRowContext(ctx, query, personID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPersonNotFound
		}
		return err
	}

	return nil
}

func (m PersonContactModel) ensureContactTypeExists(ctx context.Context, contactTypeID int) error {
	query := `
		SELECT contact_type_id
		FROM contact_type
		WHERE contact_type_id = $1
	`

	var id int
	err := m.DB.QueryRowContext(ctx, query, contactTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrContactTypeNotFound
		}
		return err
	}

	return nil
}

func (m PersonContactModel) Insert(contact *PersonContact) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.ensurePersonExists(ctx, contact.PersonID)
	if err != nil {
		return err
	}

	err = m.ensureContactTypeExists(ctx, contact.ContactTypeID)
	if err != nil {
		return err
	}

	if contact.IsPrimary {
		clearPrimaryQuery := `
			UPDATE person_contact
			SET is_primary = FALSE
			WHERE person_id = $1
			  AND is_primary = TRUE
		`
		_, err = m.DB.ExecContext(ctx, clearPrimaryQuery, contact.PersonID)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
		VALUES ($1, $2, $3, $4)
		RETURNING person_contact_id
	`

	err = m.DB.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(contact.ContactValue),
		contact.IsPrimary,
		contact.PersonID,
		contact.ContactTypeID,
	).Scan(&contact.PersonContactID)
	if err != nil {
		return err
	}

	return nil
}

func (m PersonContactModel) GetByPerson(personID int64, personContactID int64) (*PersonContact, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.ensurePersonExists(ctx, personID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			pc.person_contact_id,
			pc.person_id,
			pc.contact_type_id,
			ct.contact_type_name,
			pc.contact_value,
			pc.is_primary
		FROM person_contact pc
		JOIN contact_type ct ON ct.contact_type_id = pc.contact_type_id
		WHERE pc.person_id = $1
		  AND pc.person_contact_id = $2
	`

	var contact PersonContact
	err = m.DB.QueryRowContext(ctx, query, personID, personContactID).Scan(
		&contact.PersonContactID,
		&contact.PersonID,
		&contact.ContactTypeID,
		&contact.ContactTypeName,
		&contact.ContactValue,
		&contact.IsPrimary,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &contact, nil
}

func (m PersonContactModel) GetAllByPerson(personID int64, filters Filters) ([]*PersonContact, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.ensurePersonExists(ctx, personID)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) OVER(),
			pc.person_contact_id,
			pc.person_id,
			pc.contact_type_id,
			ct.contact_type_name,
			pc.contact_value,
			pc.is_primary
		FROM person_contact pc
		JOIN contact_type ct ON ct.contact_type_id = pc.contact_type_id
		WHERE pc.person_id = $1
		ORDER BY %s %s, pc.person_contact_id ASC
		LIMIT $2 OFFSET $3
	`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(ctx, query, personID, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	contacts := []*PersonContact{}

	for rows.Next() {
		var contact PersonContact
		err := rows.Scan(
			&totalRecords,
			&contact.PersonContactID,
			&contact.PersonID,
			&contact.ContactTypeID,
			&contact.ContactTypeName,
			&contact.ContactValue,
			&contact.IsPrimary,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		contacts = append(contacts, &contact)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return contacts, metadata, nil
}

func (m PersonContactModel) Update(contact *PersonContact) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.ensurePersonExists(ctx, contact.PersonID)
	if err != nil {
		return err
	}

	err = m.ensureContactTypeExists(ctx, contact.ContactTypeID)
	if err != nil {
		return err
	}

	if contact.IsPrimary {
		clearPrimaryQuery := `
			UPDATE person_contact
			SET is_primary = FALSE
			WHERE person_id = $1
			  AND person_contact_id <> $2
			  AND is_primary = TRUE
		`
		_, err = m.DB.ExecContext(ctx, clearPrimaryQuery, contact.PersonID, contact.PersonContactID)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE person_contact
		SET contact_value = $1,
			is_primary = $2,
			contact_type_id = $3
		WHERE person_contact_id = $4
		  AND person_id = $5
	`

	result, err := m.DB.ExecContext(
		ctx,
		query,
		strings.TrimSpace(contact.ContactValue),
		contact.IsPrimary,
		contact.ContactTypeID,
		contact.PersonContactID,
		contact.PersonID,
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

func (m PersonContactModel) Delete(personID int64, personContactID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.ensurePersonExists(ctx, personID)
	if err != nil {
		return err
	}

	query := `
		DELETE FROM person_contact
		WHERE person_contact_id = $1
		  AND person_id = $2
	`

	result, err := m.DB.ExecContext(ctx, query, personContactID, personID)
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
