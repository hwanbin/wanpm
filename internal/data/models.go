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
	Role       RoleModel
	Client     ClientModel
	Proposal   ProposalModel
	Project    ProjectModel
	User       UserModel
	Token      TokenModel
	Activity   ActivityModel
	Assignment AssignmentModel
	Timesheet  TimesheetModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Role:       RoleModel{DB: db},
		Client:     ClientModel{DB: db},
		Proposal:   ProposalModel{DB: db},
		Project:    ProjectModel{DB: db},
		User:       UserModel{DB: db},
		Token:      TokenModel{DB: db},
		Activity:   ActivityModel{DB: db},
		Assignment: AssignmentModel{DB: db},
		Timesheet:  TimesheetModel{DB: db},
	}
}
