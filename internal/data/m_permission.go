package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

// Permissions are a slice so they can be searchable and iterable
type Permissions []string

// Include method checks if a specific permission code is included in the Permissions slice
func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

// PermissionModel provides methods for working with permissions data in the database
type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser retrieves all permission codes associated with a specific user ID
func (p PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users_permissions.user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := Permissions{}
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForUser adds one or more permission codes to a specific user ID
func (p PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
		INSERT INTO users_permissions
		SELECT $1, permissions.id FROM permissions
		WHERE permissions.code = ANY($2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
