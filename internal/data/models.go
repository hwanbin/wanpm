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
	Proposal   ProposalModel
	Project    ProjectModel
	Activity   ActivityModel
	Role       RoleModel
	Token      TokenModel
	User       UserModel
	Assignment AssignmentModel
	Timesheet  TimesheetModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Client:     ClientModel{DB: db},
		Proposal:   ProposalModel{DB: db},
		Project:    ProjectModel{DB: db},
		Activity:   ActivityModel{DB: db},
		Role:       RoleModel{DB: db},
		Token:      TokenModel{DB: db},
		User:       UserModel{DB: db},
		Assignment: AssignmentModel{DB: db},
		Timesheet:  TimesheetModel{DB: db},
	}
}
