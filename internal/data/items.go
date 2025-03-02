package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoRecord     = errors.New("models: no matching record found")
	ErrInvalidInput = errors.New("models: invalid input")
	ErrEditConflict = errors.New("models: edit conflict")
)

type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Quantity  int32     `json:"quantity"`
	Remaining int32     `json:"remaining"`
	Remarks   string    `json:"remarks"`
	CreatedAt time.Time `json:"created_at"`
	Version   int32     `json:"version"`
}

type ItemModel struct {
	DB *sql.DB
}

func (m ItemModel) InsertItem(tx *sql.Tx, item *Item) error {
	query := `
		INSERT INTO items (name,  quantity, remaining, remarks)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, remaining, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return tx.QueryRowContext(ctx, query, item.Name, item.Quantity, item.Quantity, item.Remarks).Scan(
		&item.ID,
		&item.CreatedAt,
		&item.Remaining,
		&item.Version,
	)
}

func (m ItemModel) GetItem(id int64) (*Item, error) {
	if id < 1 {
		return nil, ErrNoRecord
	}

	query := `
		SELECT id, name,  quantity, remaining, remarks, created_at, version
		FROM items
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var item Item

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Quantity,
		&item.Remaining,
		&item.Remarks,
		&item.CreatedAt,
		&item.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return &item, nil
}

func (m ItemModel) GetAllItems(name string, remarks string, filters Filters) ([]*Item, Metadata, error) {
	query := `
		SELECT count(*) OVER(), id, name,  quantity, remaining, remarks, created_at, version
		FROM items
		WHERE (LOWER(name) = LOWER($1) OR $1 = '')
		AND (remarks ILIKE '%' || $2 || '%' OR $2 = '')
		ORDER BY ` + filters.sortColumn() + ` ` + filters.sortDirection() + `
		LIMIT $3 OFFSET $4`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{name, remarks, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, Metadata{}, ErrNoRecord
		default:
			return nil, Metadata{}, err
		}
	}
	defer rows.Close()

	totalRecords := 0
	items := []*Item{}

	for rows.Next() {
		var item Item

		err := rows.Scan(
			&totalRecords,
			&item.ID,
			&item.Name,
			&item.Quantity,
			&item.Remaining,
			&item.Remarks,
			&item.CreatedAt,
			&item.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return items, metadata, nil
}

func (m ItemModel) UpdateRemaining(tx *sql.Tx, id int64, removed int32, version int32) error {
	if removed < 0 {
		return ErrInvalidInput
	}

	query := `
		UPDATE items
		SET remaining = remaining - $1, version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, removed, id, version).Scan(&version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m ItemModel) DeleteItem(id int64) error {
	if id < 1 {
		return ErrNoRecord
	}

	query := `
		DELETE FROM items
		WHERE id = $1`

	_, err := m.DB.Exec(query, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNoRecord
		default:
			return err
		}
	}

	return nil
}

func (m ItemModel) UpdateItem(item *Item) error {
	if item.ID < 1 {
		return ErrNoRecord
	}

	query := `
		UPDATE items
		SET  remaining = $1, version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, item.Remaining, item.ID, item.Version).Scan(&item.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m ItemModel) AddRemaining(tx *sql.Tx, id int64, removed int32, version int32) error {
	if removed < 0 {
		return ErrInvalidInput
	}

	query := `
		UPDATE items
		SET remaining = remaining + $1, version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, removed, id, version).Scan(&version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
