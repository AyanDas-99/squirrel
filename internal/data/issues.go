package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Issue struct {
	ID       int64     `json:"id"`
	ItemID   int64     `json:"item_id"`
	Quantity int32     `json:"quantity"`
	IssuedTo string    `json:"issued_to"`
	IssuedAt time.Time `json:"issued_at"`
}

type IssueModel struct {
	DB *sql.DB
}

func (m IssueModel) InsertIssue(tx *sql.Tx, issue *Issue) error {
	query := `
		INSERT INTO issues (item_id, quantity, issued_to, issued_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, issued_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return tx.QueryRowContext(ctx, query, issue.ItemID, issue.Quantity, issue.IssuedTo, issue.IssuedAt).Scan(
		&issue.ID,
		&issue.IssuedAt,
	)
}

func (m IssueModel) GetIssues(itemID int64, filters Filters) ([]*Issue, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, item_id, quantity, issued_to, issued_at
		FROM issues
		WHERE item_id = $1
		ORDER BY %s %s
		LIMIT %d OFFSET %d`, filters.sortColumn(), filters.sortDirection(), filters.limit(), filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, itemID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, Metadata{}, ErrNoRecord
		default:
			return nil, Metadata{}, err
		}
	}
	defer rows.Close()

	issues := []*Issue{}
	totalRecords := 0

	for rows.Next() {
		var issue Issue
		err := rows.Scan(&totalRecords, &issue.ID, &issue.ItemID, &issue.Quantity, &issue.IssuedTo, &issue.IssuedAt)
		if err != nil {
			return nil, Metadata{}, err
		}
		issues = append(issues, &issue)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return issues, metadata, nil
}
