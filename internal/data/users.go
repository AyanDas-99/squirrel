package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"username"`
	Password  string    `json:"password"`
	Hash      string    `json:"-"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

type UserModel struct {
	DB *sql.DB
}

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (username, hash, is_admin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, version`

	args := []interface{}{user.UserName, user.Hash, user.IsAdmin, user.CreatedAt, user.UpdatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Version)
	if err != nil {
		switch {
		case err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"":
			return ErrDuplicateName
		}
		return err
	}

	return nil
}

func (m UserModel) GetUserByUserName(username string) (*User, error) {
	query := `
		SELECT id, username, hash, is_admin, created_at, updated_at, version
		FROM users
		WHERE username = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.UserName,
		&user.Hash,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
    UPDATE users SET username = $1, hash = $2, updated_at = CURRENT_TIMESTAMP, version = version + 1
    WHERE id = $3 AND version = $4
    RETURNING version
  `

	args := []interface{}{
		user.UserName,
		user.Hash,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateName
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(token string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(token))

	query := `
	SELECT users.id, users.username, users.hash, users.is_admin, users.created_at, users.updated_at, users.version
	FROM users
	INNER JOIN tokens ON users.id = tokens.user_id
	WHERE tokens.hash = $1 AND tokens.expiry > $2
	`

	args := []interface{}{tokenHash[:], time.Now().UTC()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.UserName,
		&user.Hash,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return &user, nil

}
