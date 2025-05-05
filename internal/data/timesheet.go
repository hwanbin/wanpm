package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hwanbin/wanpm/internal/validator"
)

var (
	ErrDuplicateTimesheetID = errors.New("duplicate timesheet id")
)

var (
	UlidRegex = regexp.MustCompile(`^[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}$`)
)

const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusCanceled  = "canceled"
	StatusSubmitted = "submitted"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
)

type TimesheetInput struct {
	UserID      *string `json:"user_id"`
	ProjectID   *int32  `json:"project_id"`
	ActivityID  *int32  `json:"activity_id"`
	ClientID    *int32  `json:"client_id"`
	WorkDate    *string `json:"work_date"`
	WorkMinutes *int32  `json:"work_mins"`
	Description *string `json:"description"`
}

type Timesheet struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ProjectID   int32     `json:"project_id"`
	ActivityID  int32     `json:"activity_id"`
	ClientID    int32     `json:"client_id"`
	WorkDate    string    `json:"work_date"`
	WorkMinutes int32     `json:"work_mins"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Version     int32     `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TimesheetClient struct {
	ID        int32      `json:"id"`
	Name      string     `json:"name"`
	Address   string     `json:"address,omitempty"`
	LogoURL   string     `json:"logo_url,omitempty"`
	Note      string     `json:"note,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type TimesheetProject struct {
	ProjectID  int32      `json:"project_id"`
	ProposalID string     `json:"proposal_id,omitempty"`
	Name       string     `json:"name"`
	Status     string     `json:"status,omitempty"`
	Feature    *Feature   `json:"feature,omitempty"`
	Images     []string   `json:"images,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type TimesheetActivity struct {
	ID        int32      `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type TimesheetUser struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Activated bool       `json:"activated"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type JoinedTimesheet struct {
	ID          string            `json:"id"`
	Client      TimesheetClient   `json:"client,omitempty"`
	Project     TimesheetProject  `json:"project,omitempty"`
	Activity    TimesheetActivity `json:"activity,omitempty"`
	User        TimesheetUser     `json:"user,omitempty"`
	WorkDate    string            `json:"work_date"`
	WorkMinutes int32             `json:"work_mins"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Version     int32             `json:"-"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type TimesheetQsInput struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	ProjectID    int32  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	ActivityID   int32  `json:"activity_id"`
	ActivityName string `json:"activity_name"`
	WorkDate     string `json:"work_date"`
	FromWorkDate string `json:"from_date"`
	ToWorkDate   string `json:"to_date"`
	Filters
}

func ValidateTimesheetInputRequired(v *validator.Validator, t *TimesheetInput) {
	v.Check(t.UserID != nil, "user_id", "must be provided")
	v.Check(t.ProjectID != nil, "project_id", "must be provided")
	v.Check(t.ClientID != nil, "client_id", "must be provided")
	v.Check(t.ActivityID != nil, "activity_id", "must be provided")
	v.Check(t.WorkDate != nil, "work_date", "must be provided")
	v.Check(t.WorkMinutes != nil, "work_mins", "must be provided")
}

func ValidateTimesheetSemantic(v *validator.Validator, t *Timesheet) {
	v.Check(t.UserID != "", "user_id", "cannot be empty")
	v.Check(len(t.UserID) == 8, "user_id", "must be 8 bytes long")
	v.Check(t.ProjectID > 0, "project_id", "must be greater than zero")
	v.Check(t.ClientID > 0, "client_id", "must be greater than zero")
	v.Check(t.ActivityID > 0, "activity_id", "must be greater than zero")
	v.Check(t.WorkDate != "", "work_date", "cannot be empty")
	v.Check(t.WorkMinutes > 0, "work_mins", "must be greater than zero")
	v.Check(t.WorkMinutes <= 1440, "work_mins", "cannot exceed 1440 minutes")
	v.Check(len(t.Description) <= 1000, "description", "cannot exceed 1000 bytes long")
}

func ValidateTotalMinutes(v *validator.Validator, timesheets []*Timesheet, workMins int32) {
	var totalMinutes int32 = 0
	if len(timesheets) > 0 {
		for _, timesheet := range timesheets {
			totalMinutes += timesheet.WorkMinutes
		}
	}
	fmt.Println(totalMinutes)
	v.Check(totalMinutes+workMins <= 1440, "work_mins", "cannot exceed 1440 minutes in total for the given date")
}

const dateLayout = "2006-01-02"

func ValidateTimesheetQueryString(v *validator.Validator, qs TimesheetQsInput) {
	if qs.WorkDate != "" {
		_, err := time.Parse(dateLayout, qs.WorkDate)
		v.Check(nil == err, "work_date", "must be a valid date format(YYYY-MM-DD)")
	}
	if qs.FromWorkDate != "" {
		_, err := time.Parse(dateLayout, qs.FromWorkDate)
		v.Check(nil == err, "from_work_date", "must be a valid date format(YYYY-MM-DD)")
	}
	if qs.ToWorkDate != "" {
		_, err := time.Parse(dateLayout, qs.ToWorkDate)
		v.Check(nil == err, "to_work_date", "must be a valid date format(YYYY-MM-DD)")
	}

	if qs.FromWorkDate != "" && qs.ToWorkDate != "" {
		from, errFrom := time.Parse(dateLayout, qs.FromWorkDate)
		v.Check(nil == errFrom, "from_date", "must be a valid date format(YYYY-MM-DD)")
		to, errTo := time.Parse(dateLayout, qs.ToWorkDate)
		v.Check(nil == errTo, "to_date", "must be a valid date format(YYYY-MM-DD)")

		if nil == errFrom && nil == errTo {
			v.Check(!from.After(to), "from_date", "cannot be after to_date")
		}
	}
}

type TimesheetModel struct {
	DB *sql.DB
}

func (m TimesheetModel) Insert(timesheet *Timesheet) error {
	query := `
		INSERT INTO timesheet (id, appuser_id, project_id, client_id, activity_id, work_date, work_minutes, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING version, created_at, updated_at`
	args := []any{
		timesheet.ID,
		timesheet.UserID,
		timesheet.ProjectID,
		timesheet.ClientID,
		timesheet.ActivityID,
		timesheet.WorkDate,
		timesheet.WorkMinutes,
		timesheet.Description,
		timesheet.Status,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&timesheet.Version,
		&timesheet.CreatedAt,
		&timesheet.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "timesheet_id_key"`:
			return ErrDuplicateTimesheetID
		default:
			return err
		}
	}

	query = `
		INSERT INTO timesheet_client (timesheet_id, client_id)
		VALUES ($1, $2)`
	args = []any{
		timesheet.ID,
		timesheet.ClientID,
	}
	result, err := tx.ExecContext(ctx, query, args...)
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

	query = `
		INSERT INTO timesheet_project (timesheet_id, project_id)
		VALUES ($1, $2)`
	args = []any{
		timesheet.ID,
		timesheet.ProjectID,
	}
	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	query = `
		INSERT INTO timesheet_appuser (timesheet_id, appuser_id)
		VALUES ($1, $2)`
	args = []any{
		timesheet.ID,
		timesheet.UserID,
	}
	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	query = `
		INSERT INTO timesheet_activity (timesheet_id, activity_id)
		VALUES ($1, $2)`
	args = []any{
		timesheet.ID,
		timesheet.ActivityID,
	}
	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m TimesheetModel) InternalGet(id string) (*Timesheet, error) {
	if id == "" {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT id, appuser_id, project_id, client_id, activity_id, work_date, work_minutes, description, status, version, created_at, updated_at
		FROM timesheet
		WHERE id = $1`
	var timesheet Timesheet

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{id}
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&timesheet.ID,
		&timesheet.UserID,
		&timesheet.ProjectID,
		&timesheet.ClientID,
		&timesheet.ActivityID,
		&timesheet.WorkDate,
		&timesheet.WorkMinutes,
		&timesheet.Description,
		&timesheet.Status,
		&timesheet.Version,
		&timesheet.CreatedAt,
		&timesheet.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &timesheet, nil
}

func (m TimesheetModel) InternalGetAll(userID, workDate string) ([]*Timesheet, error) {

	query := `
		SELECT id, appuser_id, work_date, work_minutes
		FROM timesheet t
		WHERE (
			( t.appuser_id = $1 )
			AND
			( CAST(t.work_date AS TEXT) = $2 )
		) 
		ORDER BY t.id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{userID, workDate}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timesheets []*Timesheet
	for rows.Next() {
		var timesheet Timesheet
		err = rows.Scan(
			&timesheet.ID,
			&timesheet.UserID,
			&timesheet.WorkDate,
			&timesheet.WorkMinutes,
		)
		if err != nil {
			return nil, err
		}
		timesheets = append(timesheets, &timesheet)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (m TimesheetModel) Get(id string) (*JoinedTimesheet, error) {
	query := `
		SELECT t.id, t.appuser_id,
		jsonb_build_object(
			'id', u.id,
			'email', u.email,
			'first_name', u.first_name,
			'last_name', u.last_name,
			'activated', u.activated,
			'created_at', u.created_at,
			'updated_at', u.updated_at
		) as user,
		jsonb_build_object(
			'project_id', p.project_id,
			'proposal_id', p.proposal_id,
			'name', p.name,
			'status', p.status,
			'feature', p.feature,
			'images', p.images,
			'created_at', p.created_at,
			'updated_at', p.updated_at
		) as project,
		jsonb_build_object (
			'id', c.id,
			'name', c.name,
			'address', c.address,
			'logo_url', c.logo_url,
			'note', c.note,
			'created_at', c.created_at,
			'updated_at', c.updated_at
		) as client,
		jsonb_build_object (
			'id', a.id,
			'name', a.name,
			'created_at', a.created_at,
			'updated_at', a.updated_at
		) as activity,
		t.project_id, t.client_id, t.activity_id, t.work_date, t.work_minutes, t.description, t.status, t.version, t.created_at, t.updated_at
		FROM timesheet t
		LEFT JOIN timesheet_appuser tu ON t.id = tu.timesheet_id
		LEFT JOIN appuser u ON tu.appuser_id = u.id
		LEFT JOIN timesheet_project tp ON t.id = tp.timesheet_id
		LEFT JOIN project p ON tp.project_id = p.project_id
		LEFT JOIN timesheet_client tc ON t.id = tc.timesheet_id
		LEFT JOIN client c ON tc.client_id = c.id
		LEFT JOIN timesheet_activity ta ON t.id = ta.timesheet_id
		LEFT JOIN activity a ON ta.activity_id = a.id
		WHERE t.id = $1`
	var timesheet Timesheet
	var userJSONB string
	var projectJSONB string
	var clientJSONB string
	var activityJSONB string
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		id,
	}
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&timesheet.ID,
		&timesheet.UserID,
		&userJSONB,
		&projectJSONB,
		&clientJSONB,
		&activityJSONB,
		&timesheet.ProjectID,
		&timesheet.ClientID,
		&timesheet.ActivityID,
		&timesheet.WorkDate,
		&timesheet.WorkMinutes,
		&timesheet.Description,
		&timesheet.Status,
		&timesheet.Version,
		&timesheet.CreatedAt,
		&timesheet.UpdatedAt,
	)
	fmt.Println(userJSONB)
	fmt.Println(projectJSONB)
	fmt.Println(clientJSONB)
	fmt.Println(activityJSONB)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	timesheetClient := TimesheetClient{}
	err = json.Unmarshal([]byte(clientJSONB), &timesheetClient)
	if err != nil {
		return nil, errors.New("failed to unmarshal client jsonb")
	}

	timesheetProject := TimesheetProject{}
	err = json.Unmarshal([]byte(projectJSONB), &timesheetProject)
	if err != nil {
		return nil, errors.New("failed to unmarshal project jsonb")
	}

	timesheetActivity := TimesheetActivity{}
	err = json.Unmarshal([]byte(activityJSONB), &timesheetActivity)
	if err != nil {
		return nil, errors.New("failed to unmarshal activity jsonb")
	}

	timesheetUser := TimesheetUser{}
	err = json.Unmarshal([]byte(userJSONB), &timesheetUser)
	if err != nil {
		return nil, errors.New("failed to unmarshal user jsonb")
	}

	joinedTimesheet := JoinedTimesheet{
		ID:          timesheet.ID,
		WorkDate:    timesheet.WorkDate,
		WorkMinutes: timesheet.WorkMinutes,
		Description: timesheet.Description,
		Status:      timesheet.Status,
		CreatedAt:   timesheet.CreatedAt,
		UpdatedAt:   timesheet.UpdatedAt,
		Client:      timesheetClient,
		Project:     timesheetProject,
		Activity:    timesheetActivity,
		User:        timesheetUser,
	}

	// return &timesheet, nil
	return &joinedTimesheet, nil
}

func (m TimesheetModel) GetAll(qs TimesheetQsInput) ([]*JoinedTimesheet, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), t.id,
		jsonb_build_object (
			'id', u.id,
			'email', u.email,
			'first_name', u.first_name,
			'last_name', u.last_name
		) AS user,
		jsonb_build_object (
			'project_id', p.project_id,
			'name', p.name
		) AS project,
		jsonb_build_object (
			'id', c.id,
			'name', c.name
		) AS client,
		jsonb_build_object (
			'id', a.id,
			'name', a.name
		) AS activity,
		t.work_date, t.work_minutes, t.description, t.status, t.version, t.created_at, t.updated_at
		FROM timesheet t
		LEFT JOIN timesheet_appuser tu ON t.id = tu.timesheet_id
		LEFT JOIN appuser u ON tu.appuser_id = u.id
		LEFT JOIN timesheet_project tp ON t.id = tp.timesheet_id
		LEFT JOIN project p ON tp.project_id = p.project_id
		LEFT JOIN timesheet_client tc ON t.id = tc.timesheet_id
		LEFT JOIN client c ON tc.client_id = c.id
		LEFT JOIN timesheet_activity ta ON t.id = ta.timesheet_id
		LEFT JOIN activity a ON ta.activity_id = a.id
		WHERE (
			( u.id ILIKE '%%' || $1 || '%%' OR $1 = '' )
			AND
			( u.email ILIKE '%%' || $2 || '%%' OR $2 = '' )
			AND
			( u.first_name ILIKE '%%' || $3 || '%%' OR $3 = '' )
			AND
			( u.last_name ILIKE '%%' || $4 || '%%' OR $4 = '' )
			AND
			( CAST(p.project_id AS TEXT) LIKE '%%' || $5 || '%%' OR CAST($5 AS INTEGER) = 0 )
			AND
			( p.name ILIKE '%%' || $6 || '%%' OR $6 = '' )
			AND
			( a.id = $7 OR $7 = 0 )
			AND
			( a.name ILIKE '%%' || $8 || '%%' OR $8 = '' )
			AND
			( CAST(t.work_date AS TEXT) LIKE '%%' || $9 || '%%' OR $9 = '' )
			AND
			( 
				t.work_date BETWEEN 
				CAST(COALESCE(NULLIF($10, ''), '0001-01-01') AS DATE) 
				AND 
				CAST(COALESCE(NULLIF($11, ''), '9999-12-31') AS DATE)
			)
		)
		ORDER BY %s %s, t.id ASC`, qs.Filters.sortColumn(), qs.Filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		qs.UserID,
		qs.Email,
		qs.FirstName,
		qs.LastName,
		qs.ProjectID,
		qs.ProjectName,
		qs.ActivityID,
		qs.ActivityName,
		qs.WorkDate,
		qs.FromWorkDate,
		qs.ToWorkDate,
	}

	if qs.Filters.limit() > 0 {
		query += `LIMIT $12 OFFSET $13`
		args = append(args, qs.Filters.limit(), qs.Filters.offset())
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var timesheets []*JoinedTimesheet
	for rows.Next() {
		var timesheet JoinedTimesheet
		var userJSONB string
		var projectJSONB string
		var clientJSONB string
		var activityJSONB string

		err := rows.Scan(
			&totalRecords,
			&timesheet.ID,
			&userJSONB,
			&projectJSONB,
			&clientJSONB,
			&activityJSONB,
			&timesheet.WorkDate,
			&timesheet.WorkMinutes,
			&timesheet.Description,
			&timesheet.Status,
			&timesheet.Version,
			&timesheet.CreatedAt,
			&timesheet.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, nil
		}

		err = json.Unmarshal([]byte(userJSONB), &timesheet.User)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal user json")
		}

		err = json.Unmarshal([]byte(projectJSONB), &timesheet.Project)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal project json")
		}

		err = json.Unmarshal([]byte(clientJSONB), &timesheet.Client)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal client json")
		}

		err = json.Unmarshal([]byte(activityJSONB), &timesheet.Activity)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal activity json")
		}

		timesheets = append(timesheets, &timesheet)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, qs.Filters.Page, qs.Filters.PageSize)

	return timesheets, metadata, nil
}

func (m TimesheetModel) GetAllTimesheets(qs TimesheetQsInput) ([]*JoinedTimesheet, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), t.id,
		jsonb_build_object (
			'id', u.id,
			'email', u.email,
			'first_name', u.first_name,
			'last_name', u.last_name
		) AS user,
		jsonb_build_object (
			'project_id', p.project_id,
			'name', p.name
		) AS project,
		jsonb_build_object (
			'id', c.id,
			'name', c.name
		) AS client,
		jsonb_build_object (
			'id', a.id,
			'name', a.name
		) AS activity,
		t.work_date, t.work_minutes, t.description, t.status, t.version, t.created_at, t.updated_at
		FROM timesheet t
		LEFT JOIN timesheet_appuser tu ON t.id = tu.timesheet_id
		LEFT JOIN appuser u ON tu.appuser_id = u.id
		LEFT JOIN timesheet_project tp ON t.id = tp.timesheet_id
		LEFT JOIN project p ON tp.project_id = p.project_id
		LEFT JOIN timesheet_client tc ON t.id = tc.timesheet_id
		LEFT JOIN client c ON tc.client_id = c.id
		LEFT JOIN timesheet_activity ta ON t.id = ta.timesheet_id
		LEFT JOIN activity a ON ta.activity_id = a.id
		WHERE (
			( u.id ILIKE '%%' || $1 || '%%' OR $1 = '' )
			AND
			( u.email ILIKE '%%' || $2 || '%%' OR $2 = '' )
			AND
			( u.first_name ILIKE '%%' || $3 || '%%' OR $3 = '' )
			AND
			( u.last_name ILIKE '%%' || $4 || '%%' OR $4 = '' )
			AND
			( CAST(p.project_id AS TEXT) LIKE '%%' || $5 || '%%' OR CAST($5 AS INTEGER) = 0 )
			AND
			( p.name ILIKE '%%' || $6 || '%%' OR $6 = '' )
			AND
			( CAST(a.id AS TEXT) LIKE '%%' || $7 || '%%' OR CAST($7 AS INTEGER) = 0 )
			AND
			( a.name ILIKE '%%' || $8 || '%%' OR $8 = '' )
		)
		ORDER BY %s %s, t.id ASC`, qs.Filters.sortColumn(), qs.Filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		qs.UserID,
		qs.Email,
		qs.FirstName,
		qs.LastName,
		qs.ProjectID,
		qs.ProjectName,
		qs.ActivityID,
		qs.ActivityName,
		// qs.WorkDate,
		// qs.FromWorkDate,
		// qs.ToWorkDate,
	}

	// if qs.Filters.limit() > 0 {
	// 	query += `LIMIT $12 OFFSET $13`
	// 	args = append(args, qs.Filters.limit(), qs.Filters.offset())
	// }
	fmt.Println(query)

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var timesheets []*JoinedTimesheet
	for rows.Next() {
		var timesheet JoinedTimesheet
		var userJSONB string
		var projectJSONB string
		var clientJSONB string
		var activityJSONB string

		err := rows.Scan(
			&totalRecords,
			&timesheet.ID,
			&userJSONB,
			&projectJSONB,
			&clientJSONB,
			&activityJSONB,
			&timesheet.WorkDate,
			&timesheet.WorkMinutes,
			&timesheet.Description,
			&timesheet.Status,
			&timesheet.Version,
			&timesheet.CreatedAt,
			&timesheet.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, nil
		}

		err = json.Unmarshal([]byte(userJSONB), &timesheet.User)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal user json")
		}

		err = json.Unmarshal([]byte(projectJSONB), &timesheet.Project)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal project json")
		}

		err = json.Unmarshal([]byte(clientJSONB), &timesheet.Client)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal client json")
		}

		err = json.Unmarshal([]byte(activityJSONB), &timesheet.Activity)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal activity json")
		}

		timesheets = append(timesheets, &timesheet)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, qs.Filters.Page, qs.Filters.PageSize)

	return timesheets, metadata, nil
}

func (m TimesheetModel) Update(timesheet *Timesheet) error {
	query := `
		UPDATE timesheet
		SET appuser_id = $1, project_id = $2, client_id = $3, activity_id = $4, work_date = $5, work_minutes = $6, description = $7, version = version + 1, updated_at = $8
		WHERE id = $9 AND version = $10
		RETURNING version`

	args := []any{
		timesheet.UserID,
		timesheet.ProjectID,
		timesheet.ClientID,
		timesheet.ActivityID,
		timesheet.WorkDate,
		timesheet.WorkMinutes,
		timesheet.Description,
		time.Now(),
		timesheet.ID,
		timesheet.Version,
	}

	fmt.Println(timesheet)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&timesheet.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	// updating timesheet_project table
	// TODO: check project_id is associated with the user - look up assignment table
	query = `
		UPDATE timesheet_project
		SET project_id = $1
		WHERE timesheet_id = $2`
	args = []any{
		timesheet.ProjectID,
		timesheet.ID,
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// updating timesheet_client table
	query = `
		UPDATE timesheet_client
		SET client_id = $1
		WHERE timesheet_id = $2`
	args = []any{
		timesheet.ClientID,
		timesheet.ID,
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// updating timesheet_activity table
	query = `
		UPDATE timesheet_activity
		SET activity_id = $1
		WHERE timesheet_id = $2`
	args = []any{
		timesheet.ActivityID,
		timesheet.ID,
	}
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m TimesheetModel) Delete(id string) error {
	if id == "" {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM timesheet
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
