package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ContactType struct {
	ContactTypeID   int    `json:"contact_type_id"`
	ContactTypeName string `json:"contact_type_name"`
}

type ContactTypeModel struct {
	DB *sql.DB
}

func (m ContactTypeModel) Get(id int) (*ContactType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT contact_type_id, contact_type_name
		FROM contact_type
		WHERE contact_type_id = $1
	`

	var contactType ContactType

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&contactType.ContactTypeID,
		&contactType.ContactTypeName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &contactType, nil
}

func (m ContactTypeModel) GetAll(filters Filters) ([]*ContactType, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), contact_type_id, contact_type_name
		FROM contact_type
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	contactTypes := []*ContactType{}

	for rows.Next() {
		var contactType ContactType

		err := rows.Scan(
			&totalRecords,
			&contactType.ContactTypeID,
			&contactType.ContactTypeName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		contactTypes = append(contactTypes, &contactType)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return contactTypes, metadata, nil
}
