package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/hwanbin/wanpm-api/internal/validator"
)

type Assignment struct {
	ProjectID  int32  `json:"project_id"`
	EmployeeID string `json:"employee_id"`
	RoleID     int32  `json:"role_id"`
}

type AssignmentModel struct {
	DB *sql.DB
}

func ValidateUserID(v *validator.Validator, assignments []*Assignment, inputUserID string) {
	for _, assignment := range assignments {
		if assignment.EmployeeID == inputUserID {
			return
		}
	}
	v.Check(false, "employee id", "not assigned to the project")
}

func (m AssignmentModel) GetAllByEmployeeID(employee_id string) ([]*Assignment, error) {
	query := `
		SELECT project_id, employee_id, role_id
		FROM assignment
		WHERE employee_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		employee_id,
	}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []*Assignment{}
	for rows.Next() {
		var assignment Assignment

		err := rows.Scan(
			&assignment.ProjectID,
			&assignment.EmployeeID,
			&assignment.RoleID,
		)
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, &assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (m AssignmentModel) GetAllByProjectID(project_id int32) ([]*Assignment, error) {
	query := `
		SELECT project_id, employee_id, role_id
		FROM assignment
		WHERE project_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		project_id,
	}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []*Assignment{}
	for rows.Next() {
		var assignment Assignment

		err := rows.Scan(
			&assignment.ProjectID,
			&assignment.EmployeeID,
			&assignment.RoleID,
		)
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, &assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}
