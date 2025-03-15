package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPermissionDoesNotExist = errors.New("permission does not exist")
	ErrDuplicatePermission    = errors.New("user already has permission")
	ErrUserDoesNotExist       = errors.New("user does not exist")
)

type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}

type PermissionModel struct {
	DB *sql.DB
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON permissions.id = users_permissions.permission_id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, code int) error {
	query := `
		INSERT INTO users_permissions (user_id, permission_id)
		VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, code)

	if err != nil {
		switch {
		case err.Error() == "pq: insert or update on table \"users_permissions\" violates foreign key constraint \"users_permissions_permission_id_fkey\"":
			return ErrPermissionDoesNotExist
		case err.Error() == "pq: duplicate key value violates unique constraint \"users_permissions_pkey\"":
			return ErrDuplicatePermission
		case err.Error() == "pq: insert or update on table \"users_permissions\" violates foreign key constraint \"users_permissions_user_id_fkey\"":
			return ErrUserDoesNotExist
		}

		return err
	}
	return err
}

func (m PermissionModel) RemoveForUser(userID int64, code int) error {
	query := `
		DELETE FROM users_permissions
		WHERE user_id = $1 AND permission_id = $2	
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, code)
	return err
}
