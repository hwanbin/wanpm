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
	Client   ClientModel
	Proposal ProposalModel
	Project  ProjectModel
	Activity ActivityModel
	Role     RoleModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Client:   ClientModel{DB: db},
		Proposal: ProposalModel{DB: db},
		Project:  ProjectModel{DB: db},
		Activity: ActivityModel{DB: db},
		Role:     RoleModel{DB: db},
	}
}
