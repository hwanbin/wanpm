package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hwanbin/wanpm/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateID    = errors.New("duplicate id")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int32     `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserQsInput struct {
	Email     string
	FirstName string
	LastName  string
	Filters
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.FirstName != "", "first name", "must be provided")
	v.Check(len(user.FirstName) <= 500, "first name", "must not be more than 500 bytes long")

	v.Check(user.LastName != "", "last name", "must be provided")
	v.Check(len(user.LastName) <= 500, "last name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO appuser (id, email, first_name, last_name, password_hash, activated)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING version, created_at, updated_at`

	args := []any{user.ID, user.Email, user.FirstName, user.LastName, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "appuser_id_key"`:
			return ErrDuplicateID
		case err.Error() == `pq: duplicate key value violates unique constraint "appuser_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, updated_at, first_name, last_name, email, password_hash, activated, version
		FROM appuser
		WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetById(id string) (*User, error) {
	query := `
		SELECT id, created_at, updated_at, first_name, last_name, email, activated
		FROM appuser
		WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Activated,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetAll(qs UserQsInput) ([]*User, Metadata, error) {

	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, updated_at, first_name, last_name, email, activated
		FROM appuser
		WHERE (
			( 
				( email ILIKE '%%' || $1 || '%%' AND NOT $1 = '' ) 
				OR
				( first_name ILIKE '%%' || $2 || '%%' AND NOT $2 = '' )
				OR
				( last_name ILIKE '%%' || $3 || '%%' AND NOT $3 = '' )
			)
			OR
			( $1 = '' AND $2 = '' AND $3 = '' )
		)
		ORDER BY %s %s, id ASC`, qs.Filters.sortColumn(), qs.Filters.sortDirection())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		qs.Email,
		qs.FirstName,
		qs.LastName,
	}

	if qs.Filters.limit() > 0 {
		query += `
			LIMIT $4 OFFSET $5`
		args = append(args, qs.Filters.limit(), qs.Filters.offset())
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var users []*User
	for rows.Next() {
		var user User

		err := rows.Scan(
			&totalRecords,
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Activated,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, qs.Filters.Page, qs.Filters.PageSize)

	return users, metadata, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE appuser
		SET first_name = $1, last_name = $2, email = $3, password_hash = $4, activated = $5, version = version + 1, updated_at = now()
		WHERE id = $6 AND version = $7
		RETURNING version`

	args := []any{
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "appuser_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT appuser.id, appuser.first_name, appuser.last_name, appuser.email, appuser.password_hash, appuser.activated, appuser.version, appuser.created_at, appuser.updated_at
        FROM appuser
        INNER JOIN token
        ON appuser.id = token.appuser_id
        WHERE token.hash = $1
        AND token.scope = $2
        AND token.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
