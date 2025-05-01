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
	ErrDuplicateActivityName = errors.New("duplicate activity name")
)

type ActivityInput struct {
	Name *string `json:"name"`
}

type Activity struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Version   int32     `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ActivityQsInput struct {
	Name string
	Filters
}

func ValidateActivityInputRequired(v *validator.Validator, a *ActivityInput) {
	v.Check(a.Name != nil, "name", "must be provided")
}

func ValidateActivityInputSemantic(v *validator.Validator, a *ActivityInput) {
	v.Check(len(*a.Name) != 0, "name", "cannot be empty")
	v.Check(len(*a.Name) <= 100, "name", "cannot be more than 100 bytes long")
}

type ActivityModel struct {
	DB *sql.DB
}

func (m ActivityModel) Insert(activity *Activity) error {
	query := `
		INSERT INTO activity (name)
		VALUES ($1)
		RETURNING id, version, created_at, updated_at`
	args := []any{
		activity.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&activity.ID,
		&activity.Version,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "activity_name_key"`:
			return ErrDuplicateActivityName
		default:
			return err
		}
	}

	return nil
}

func (m ActivityModel) Get(id int32) (*Activity, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, version, created_at, updated_at
		FROM activity
		WHERE id = $1`
	var activity Activity

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&activity.ID,
		&activity.Name,
		&activity.Version,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &activity, nil
}

func (m ActivityModel) GetAll(qs ActivityQsInput) ([]*Activity, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, version, created_at, updated_at
		FROM activity
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
	activities := []*Activity{}
	for rows.Next() {
		var activity Activity

		err := rows.Scan(
			&totalRecords,
			&activity.ID,
			&activity.Name,
			&activity.Version,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metaData := calculateMetadata(totalRecords, qs.Filters.Page, qs.Filters.PageSize)

	return activities, metaData, nil
}

func (m ActivityModel) Update(activity *Activity) error {
	query := `
		UPDATE activity
		SET name = $1, version = version + 1, updated_at = $2
		WHERE id = $3 AND version = $4
		RETURNING version, updated_at`

	args := []any{
		activity.Name,
		time.Now(),
		activity.ID,
		activity.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&activity.Version,
		&activity.UpdatedAt,
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

func (m ActivityModel) Delete(id int32) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM activity
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
