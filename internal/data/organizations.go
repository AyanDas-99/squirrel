package data

import (
	"database/sql"
	"errors"
	"time"
)

type Organization struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type OrganizationsModel struct {
	DB *sql.DB
}

func (m OrganizationsModel) InsertOrganization(org *Organization) error {
	query := `INSERT INTO organizations (name, created_at) VALUES ($1, $2) RETURNING id, created_at`

	args := []interface{}{org.Name, org.CreatedAt}

	err := m.DB.QueryRow(query, args...).Scan(&org.ID, &org.CreatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "organizations_name_key"`:
			return ErrDuplicateName
		default:
			return err
		}
	}
	return nil
}

func (m OrganizationsModel) GetOrganizations() ([]*Organization, error) {
	query := `SELECT id, name, created_at FROM organizations`

	rows, err := m.DB.Query(query)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}
	defer rows.Close()

	var orgs []*Organization

	for rows.Next() {
		var org Organization
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (m OrganizationsModel) GetOrganizationByID(id int64) (*Organization, error) {
	query := `SELECT id, name, created_at FROM organizations WHERE id = $1`

	var org Organization
	err := m.DB.QueryRow(query, id).Scan(&org.ID, &org.Name, &org.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return &org, nil
}
