package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Removal struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	Quantity  int32     `json:"quantity"`
	Remarks   string    `json:"remarks"`
	RemovedAt time.Time `json:"removed_at"`
}

type RemovalModel struct {
	DB *sql.DB
}

func (m RemovalModel) InsertRemoval(tx *sql.Tx, removal *Removal) error {
	query := `
		INSERT INTO removals (item_id, quantity, remarks)
		VALUES ($1, $2, $3)
		RETURNING id, removed_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{removal.ItemID, removal.Quantity, removal.Remarks}

	return tx.QueryRowContext(ctx, query, args...).Scan(&removal.ID, &removal.RemovedAt)
}

func (m RemovalModel) GetRemovals(itemID int64, filters Filters) ([]*Removal, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, item_id, quantity, remarks, removed_at
		FROM removals
		WHERE item_id = $1
		ORDER BY %s %s
		LIMIT %d OFFSET %d`, filters.sortColumn(), filters.sortDirection(), filters.limit(), filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, itemID)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, Metadata{}, err
	}
	defer rows.Close()

	removals := []*Removal{}
	totalRecords := 0

	for rows.Next() {
		var removal Removal

		err := rows.Scan(
			&totalRecords,
			&removal.ID,
			&removal.ItemID,
			&removal.Quantity,
			&removal.Remarks,
			&removal.RemovedAt,
		)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return nil, Metadata{}, err
		}

		removals = append(removals, &removal)
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("Error: %v", err)
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return removals, metadata, nil
}
