package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hwanbin/wanpm-api/internal/validator"
)

var (
	ErrDuplicateRoleName = errors.New("duplicate role name")
)

type RoleInput struct {
	Name *string `json:"name"`
}

type Role struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Version   int32     `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleQsInput struct {
	Name string
	Filters
}

func ValidateRoleInputRequired(v *validator.Validator, r *RoleInput) {
	v.Check(r.Name != nil, "name", "must be provided")
}

func ValidateRoleInputSemantic(v *validator.Validator, r *RoleInput) {
	v.Check(len(*r.Name) != 0, "name", "cannot be empty")
	v.Check(len(*r.Name) <= 30, "name", "cannot be more than 30 bytes long")
}

type RoleModel struct {
	DB *sql.DB
}

func (m RoleModel) Insert(role *Role) error {
	query := `
		INSERT INTO role (name)
		VALUES ($1)
		RETURNING id, version, created_at, updated_at`
	args := []any{
		role.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&role.ID,
		&role.Version,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "role_name_key"`:
			return ErrDuplicateActivityName
		default:
			return err
		}
	}

	return nil
}

func (m RoleModel) Get(id int32) (*Role, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, version, created_at, updated_at
		FROM role
		WHERE id = $1`
	var role Role

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Version,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}

func (m RoleModel) GetAll(qs RoleQsInput) ([]*Role, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, version, created_at, updated_at
		FROM role
		WHERE ( name ILIKE '%%' || $1 || '%%' OR $1 = '' )
		ORDER BY %s %s, id ASC`, qs.Filters.sortColumn(), qs.Filters.sortDirection())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		qs.Name,
	}

	if qs.Filters.limit() > 0 {
		query += `
			LIMIT $2 OFFSET $3`
		args = append(args, qs.Filters.limit(), qs.Filters.offset())
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	roles := []*Role{}
	for rows.Next() {
		var role Role

		err := rows.Scan(
			&totalRecords,
			&role.ID,
			&role.Name,
			&role.Version,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		roles = append(roles, &role)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metaData := calculateMetadata(totalRecords, qs.Filters.Page, qs.Filters.PageSize)

	return roles, metaData, nil
}

func (m RoleModel) Update(role *Role) error {
	query := `
		UPDATE role
		SET name = $1, version = version + 1, updated_at = $2
		WHERE id = $3 AND version = $4
		RETURNING version, updated_at`

	args := []any{
		role.Name,
		time.Now(),
		role.ID,
		role.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&role.Version,
		&role.UpdatedAt,
	)
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

func (m RoleModel) Delete(id int32) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM role
		WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
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
