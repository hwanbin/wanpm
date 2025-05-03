package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hwanbin/wanpm/internal/validator"
)

type Client struct {
	InternalID int32     `json:"id"`
	Name       *string   `json:"name"`
	Address    *string   `json:"address"`
	LogoURL    *string   `json:"logo_url"`
	Note       *string   `json:"note"`
	Version    int32     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ValidateClient(v *validator.Validator, client *Client) {
	v.Check(*client.Name != "", "name", "must be provided")
	v.Check(len(*client.Name) <= 500, "name", "must not be more than 500 bytes long")

	if client.Address != nil {
		v.Check(*client.Address != "", "address", "must not be empty string")
		v.Check(len(*client.Address) <= 500, "address", "must not be more than 500 bytes long")
	}

	if client.Note != nil {
		v.Check(*client.Note != "", "contact_info", "must not be empty string")
		v.Check(len(*client.Note) <= 500, "contact_info", "must not be more than 500 bytes long")
	}
}

type ClientModel struct {
	DB *sql.DB
}

func (m ClientModel) Insert(client *Client) error {
	query := `
		INSERT INTO client (name, address, logo_url, note)
		VALUES ($1, $2, $3, $4)
		RETURNING internal_id, version, created_at, updated_at`

	args := []any{client.Name, client.Address, client.LogoURL, client.Note}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
		&client.InternalID,
		&client.Version,
		&client.CreatedAt,
		&client.UpdatedAt,
	)
}

func (m ClientModel) Get(internal_id int32) (*Client, error) {
	if internal_id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT internal_id, name, address, logo_url, note, version, created_at, updated_at
		FROM client
		WHERE internal_id = $1`
	var client Client

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, internal_id).Scan(
		&client.InternalID,
		&client.Name,
		&client.Address,
		&client.LogoURL,
		&client.Note,
		&client.Version,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &client, nil
}

func (m ClientModel) GetAll(name string, filters Filters) ([]*Client, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), internal_id, name, address, logo_url, note, version, created_at, updated_at
		FROM client
		WHERE ( to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, internal_id ASC`, filters.sortColumn(), filters.sortDirection())

	args := []any{
		name,
	}

	if filters.limit() > 0 {
		query += `
		LIMIT $2 OFFSET $3`
		args = append(args, filters.limit(), filters.offset())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	clients := []*Client{}

	for rows.Next() {
		var client Client
		err := rows.Scan(
			&totalRecords,
			&client.InternalID,
			&client.Name,
			&client.Address,
			&client.LogoURL,
			&client.Note,
			&client.Version,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		clients = append(clients, &client)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return clients, metadata, nil
}

func (m ClientModel) GetClientByName(name string) (*Client, error) {
	if name == "" {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT internal_id, name, address, logo_url, note, version, created_at, updated_at
		FROM client
		WHERE name = $1`

	var client Client

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, name).Scan(
		&client.InternalID,
		&client.Name,
		&client.Address,
		&client.LogoURL,
		&client.Note,
		&client.Version,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &client, nil
}

func (cm ClientModel) Update(c *Client) error {
	query := `
		UPDATE client
		SET name = $1, address = $2, logo_url = $3, note = $4
		WHERE internal_id = $5
		RETURNING version`

	args := []any{
		c.Name,
		c.Address,
		c.LogoURL,
		c.Note,
		c.InternalID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return cm.DB.QueryRowContext(ctx, query, args...).Scan(&c.Version)
}

func (cm ClientModel) Delete(internal_id int32) error {
	if internal_id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM client
		WHERE internal_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := cm.DB.ExecContext(ctx, query, internal_id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
