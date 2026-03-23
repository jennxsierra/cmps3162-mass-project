package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CancellationReason struct {
	ReasonID   int    `json:"reason_id"`
	ReasonName string `json:"reason_name"`
}

type CancellationReasonModel struct {
	DB *sql.DB
}

func (m CancellationReasonModel) Get(id int) (*CancellationReason, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT reason_id, reason_name
		FROM cancellation_reason
		WHERE reason_id = $1
	`

	var reason CancellationReason

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&reason.ReasonID,
		&reason.ReasonName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &reason, nil
}

func (m CancellationReasonModel) GetAll(filters Filters) ([]*CancellationReason, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), reason_id, reason_name
		FROM cancellation_reason
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, filters.sortColumn(), filters.sortDirection())

	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	reasons := []*CancellationReason{}

	for rows.Next() {
		var reason CancellationReason

		err := rows.Scan(
			&totalRecords,
			&reason.ReasonID,
			&reason.ReasonName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		reasons = append(reasons, &reason)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return reasons, metadata, nil
}
