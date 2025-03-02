package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Addition struct {
	ID       int64     `json:"id"`
	ItemID   int64     `json:"item_id"`
	Quantity int32     `json:"quantity"`
	Remarks  string    `json:"remarks"`
	AddedAt  time.Time `json:"added_at"`
}

type AdditionModel struct {
	DB *sql.DB
}

func (m AdditionModel) InsertAddition(tx *sql.Tx, addition *Addition) error {
	ctx := context.Background()
	query := `
		INSERT INTO additions (item_id, quantity, remarks)
		VALUES ($1, $2, $3)
		RETURNING id, added_at
	`
	return tx.QueryRowContext(ctx, query, addition.ItemID, addition.Quantity, addition.Remarks).Scan(&addition.ID, &addition.AddedAt)
}

func (m AdditionModel) GetAdditions(itemID int64, filters Filters) ([]*Addition, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, item_id, quantity, remarks, added_at
		FROM additions
		WHERE item_id = $1
		ORDER BY %s %s
		LIMIT %d OFFSET %d`, filters.sortColumn(), filters.sortDirection(), filters.limit(), filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	additions := []*Addition{}
	totalRecords := 0

	for rows.Next() {
		var addition Addition
		err := rows.Scan(
			&totalRecords,
			&addition.ID,
			&addition.ItemID,
			&addition.Quantity,
			&addition.Remarks,
			&addition.AddedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		additions = append(additions, &addition)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return additions, calculateMetadata(totalRecords, filters.Page, filters.PageSize), nil
}
