package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/hwanbin/wanpm-api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var AnonymousUser = &AppUser{}

type AppUser struct {
	InternalID int32     `json:"-"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activated  bool      `json:"activated"`
	Version    int       `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (u *AppUser) IsAnonymous() bool {
	return u == AnonymousUser
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

func ValidateAppUser(v *validator.Validator, user *AppUser) {
	v.Check(user.FirstName != "", "first_name", "must be provided")
	v.Check(len(user.FirstName) <= 255, "first_name", "must not be more than 255 bytes long")

	v.Check(user.LastName != "", "last_name", "must be provided")
	v.Check(len(user.LastName) <= 255, "last_name", "must not be more than 255 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type AppUserModel struct {
	DB *sql.DB
}

func (m AppUserModel) Insert(user *AppUser) error {
	query := `
		INSERT INTO appuser (first_name, last_name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING internal_id, version, created_at, updated_at`

	args := []any{user.FirstName, user.LastName, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.InternalID,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "appuser_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m AppUserModel) GetByEmail(email string) (*AppUser, error) {
	query := `
		SELECT internal_id, first_name, last_name, email, password_hash, activated, version, created_at, updated_at
		FROM appuser
		WHERE email = $1`

	var user AppUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.InternalID,
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

func (m AppUserModel) Update(user *AppUser) error {
	query := `
		UPDATE appuser
		SET first_name = $1, last_name = $2, email = $3, password_hash = $4, activated = $5, version = version + 1
		WHERE internal_id = $6 AND version = $7
		RETURNING version`

	args := []any{
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.InternalID,
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

func (m AppUserModel) GetForToken(tokenScope, tokenPlaintext string) (*AppUser, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT u.internal_id, u.first_name, u.last_name, u.email, u.password_hash, u.activated, u.version, u.created_at, u.updated_at
		FROM appuser u
		JOIN token t
		ON u.internal_id = t.appuser_internal_id
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user AppUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.InternalID,
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
