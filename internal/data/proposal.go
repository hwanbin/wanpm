package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hwanbin/wanpm-api/internal/validator"
)

type Proposal struct {
	InternalID int32     `json:"-"`
	ExternalID string    `json:"proposal_id"`
	Version    int32     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ValidateProposal(v *validator.Validator, proposal *Proposal) {
	v.Check(proposal.ExternalID != "", "proposal_id", "must be provided")
	v.Check(len(proposal.ExternalID) <= 10, "proposal_id", "must not be more than 10 bytes long")
}

type ProposalModel struct {
	DB *sql.DB
}

func (ppm ProposalModel) Insert(proposal *Proposal) error {
	query := `
		INSERT INTO proposal (project_id)
		VALUES ($1)
		RETURNING internal_id, project_id, version, created_at, updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return ppm.DB.QueryRowContext(ctx, query, proposal.ExternalID).Scan(
		&proposal.InternalID,
		&proposal.ExternalID,
		&proposal.Version,
		&proposal.CreatedAt,
		&proposal.UpdatedAt,
	)
}

func (ppm ProposalModel) Get(externalID string) (*Proposal, error) {
	if externalID == "" {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT internal_id, project_id, version, created_at, updated_at
		FROM proposal
		WHERE project_id = $1`
	var proposal Proposal

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := ppm.DB.QueryRowContext(ctx, query, externalID).Scan(
		&proposal.InternalID,
		&proposal.ExternalID,
		&proposal.Version,
		&proposal.CreatedAt,
		&proposal.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &proposal, nil
}

func (ppm ProposalModel) Update(proposal *Proposal) error {
	query := `
		UPDATE proposal
		SET project_id = $1, version = version + 1
		WHERE internal_id = $2 AND version = $3
		RETURNING version`
	args := []any{
		proposal.ExternalID,
		proposal.InternalID,
		proposal.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := ppm.DB.QueryRowContext(ctx, query, args...).Scan(
		&proposal.Version,
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

func (ppm ProposalModel) Delete(externalID string) error {
	if externalID == "" {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM proposal
		WHERE project_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := ppm.DB.ExecContext(ctx, query, externalID)
	if err != nil {
		return err
	}

	rowsAffcted, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffcted == 0 {
		return ErrRecordNotFound
	}

	return nil
}
