package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrEditConflict    = errors.New("edit conflict")
	ErrZeroRowInserted = errors.New("no row inserted")
)

type Models struct {
	Client     ClientModel
	AppUser    AppUserModel
	Proposal   ProposalModel
	Project    ProjectModel
	Token      TokenModel
	Permission PermissionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Client:     ClientModel{DB: db},
		AppUser:    AppUserModel{DB: db},
		Proposal:   ProposalModel{DB: db},
		Project:    ProjectModel{DB: db},
		Token:      TokenModel{DB: db},
		Permission: PermissionModel{DB: db},
	}
}
