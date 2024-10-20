package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hwanbin/wanpm-api/internal/validator"
	"github.com/lib/pq"
)

var (
	ErrDuplicateProjectID  = errors.New("duplicate project_id")
	ErrDuplicateProposalID = errors.New("duplicate proposal_id")
	ErrInvalidBBoxLength   = errors.New("invalid bbox length")
)

type Point struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type ProjectClient struct {
	ClientID      *int32  `json:"id"`
	ClientName    *string `json:"name"`
	ClientLogo    *string `json:"logo_url"`
	ClientAddress *string `json:"address"`
	ClientNote    *string `json:"note"`
}

type ProjectInput struct {
	ExternalID  *int32    `json:"project_id"`
	ProposalID  *string   `json:"proposal_id"`
	Name        *string   `json:"name"`
	Status      *string   `json:"status"`
	Coordinates []float64 `json:"coordinates"`
	Images      []string  `json:"images"`
	ClientNames []string  `json:"client_names"`
}

type Project struct {
	InternalID  int32           `json:"-"`
	ExternalID  *int32          `json:"project_id"`
	ProposalID  *string         `json:"proposal_id"`
	Name        *string         `json:"name"`
	Status      *string         `json:"status"`
	Coordinates []float64       `json:"coordinates"`
	Images      []string        `json:"images"`
	Clients     []ProjectClient `json:"clients"`
	Version     int32           `json:"version"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ProjectQsInput struct {
	Name    string
	Status  string
	Clients []string
	Bbox    []string
	Filters
}

func ValidateQueryString(v *validator.Validator, qs *ProjectQsInput) {
	if qs.Bbox != nil {
		v.Check(len(qs.Bbox) == 4, "bbox", "must have 4 coordinates")
	}
}

func ValidateProjectInputRequired(v *validator.Validator, p *ProjectInput) {
	v.Check(p.ExternalID != nil, "project_id", "must be provided")
	v.Check(p.ProposalID != nil, "proposal_id", "must be provided")
	v.Check(p.Name != nil, "name", "must be provided")
	v.Check(p.Status != nil, "status", "must be provided")
	v.Check(p.ClientNames != nil, "client_names", "must be provided")
}

func ValidateProjectInputSemantic(v *validator.Validator, p *ProjectInput) {
	v.Check(*p.ExternalID > 0, "project_id", "must be a positive integer")

	v.Check(*p.ProposalID != "", "proposal_id", "must not be empty string")
	v.Check(len(*p.ProposalID) <= 10, "proposal_id", "must not be more than 10 bytes long")

	v.Check(*p.Name != "", "name", "must not be empty string")
	v.Check(len(*p.Name) <= 500, "name", "must not be more than 500 bytes long")

	v.Check(*p.Status != "", "status", "must not be empty string")
	v.Check(len(*p.Status) <= 100, "status", "must not be more than 100 bytes long")

	v.Check(len(p.ClientNames) >= 1, "client_names", "must contain at least 1 client")
	v.Check(validator.Unique(p.ClientNames), "client_names", "must not contain duplicate values")

	if p.Coordinates != nil {
		v.Check(len(p.Coordinates) == 2, "coordinates", "must be a longitude, latitude pair")
		if len(p.Coordinates) == 2 {
			v.Check(p.Coordinates[0] >= -180 && p.Coordinates[0] <= 180, "coordinates", "longitude must be in between -180 and 180")
			v.Check(p.Coordinates[1] >= -90 && p.Coordinates[1] <= 90, "coordinates", "latitude must be in between -90 and 90")
		}
	}

	if p.Images != nil {
		v.Check(len(p.Images) >= 1, "images", "must contain at least 1 image")
		v.Check(validator.Unique(p.Images), "images", "must not contain duplicate values")
	}
}

func ValidateUpdatingProjectInput(v *validator.Validator, p *ProjectInput) {
	if p.ExternalID != nil {
		v.Check(*p.ExternalID > 0, "project_id", "must be a positive integer")
	}

	if p.ProposalID != nil {
		v.Check(*p.ProposalID != "", "proposal_id", "must note be empty string")
		v.Check(len(*p.ProposalID) <= 10, "proposal_id", "must not be more than 10 bytes long")
	}

	if p.Name != nil {
		v.Check(*p.Name != "", "name", "must not be empty string")
		v.Check(len(*p.Name) <= 500, "name", "must not be more than 500 bytes long")
	}

	if p.Status != nil {
		v.Check(*p.Status != "", "status", "must not be empty string")
		v.Check(len(*p.Status) <= 100, "status", "must not be more than 100 bytes long")
	}

	if p.ClientNames != nil {
		v.Check(len(p.ClientNames) >= 1, "client_names", "must contain at least 1 client")
		v.Check(validator.Unique(p.ClientNames), "client_names", "must not contain duplicate values")
	}

	if len(p.Coordinates) > 0 {
		v.Check(len(p.Coordinates) == 2, "coordinates", "must be a longitude, latitude pair")
		if len(p.Coordinates) == 2 {
			v.Check(p.Coordinates[0] >= -180 && p.Coordinates[0] <= 180, "coordinates", "longitude must be in between -180 and 180")
			v.Check(p.Coordinates[1] >= -90 && p.Coordinates[1] <= 90, "coordinates", "latitude must be in between -90 and 90")
		}
	}

	if len(p.Images) > 0 {
		v.Check(validator.Unique(p.Images), "images", "must not contain duplicate values")
	}
}

func ValidateProject(v *validator.Validator, p *Project) {
	v.Check(p.ProposalID != nil, "proposal_id", "must be provided")
	if p.ProposalID != nil {
		v.Check(*p.ProposalID != "", "proposal_id", "must not be empty string")
		v.Check(len(*p.ProposalID) <= 10, "proposal_id", "must not be more than 10 bytes long")
	}

	v.Check(p.ExternalID != nil, "project_id", "must be provided")
	if p.ExternalID != nil {
		v.Check(*p.ExternalID != 0, "project_id", "must be non-zero value")
		v.Check(*p.ExternalID > 0, "project_id", "must be a positive integer")
	}

	v.Check(p.Name != nil, "name", "must be provided")
	if p.Name != nil {
		v.Check(*p.Name != "", "name", "must not be empty string")
		v.Check(len(*p.Name) <= 500, "name", "must not be more than 500 bytes long")
	}

	v.Check(p.Status != nil, "status", "must be provided")
	if p.Status != nil {
		v.Check(*p.Status != "", "status", "must not be empty string")
		v.Check(len(*p.Status) <= 100, "status", "must not be more than 100 bytes long")
	}

	// v.Check(p.Coordinates != nil, "coordinates", "must be provided")
	if p.Coordinates != nil {
		// when below compare equals false then put it int the error bag
		v.Check(len(p.Coordinates) == 2, "coordinates", "must be a longitude, latitude pair")
	}

	if p.Images != nil {
		v.Check(len(p.Images) >= 1, "images", "must contain at least 1 image")
		v.Check(validator.Unique(p.Images), "images", "must not contain duplicate values")
	}

	v.Check(p.Clients != nil, "client_names", "must be provided")
	if p.Clients != nil {
		v.Check(len(p.Clients) >= 1, "client_names", "must contain at least 1 client")
		v.Check(validator.Unique(p.Clients), "client_names", "must not contain duplicate values")
	}
}

type ProjectModel struct {
	DB *sql.DB
}

func (m ProjectModel) Insert(project *Project) error {
	geomText := "POINT EMPTY"
	if project.Coordinates != nil {
		geomText = fmt.Sprintf("POINT(%f %f)", project.Coordinates[0], project.Coordinates[1])
	}
	query := `
		INSERT INTO project (project_id, proposal_id, name, status, coordinates, images)
		VALUES ($1, $2, $3, $4, ST_GeomFromText($5, 4326), $6)
		RETURNING internal_id, version, created_at, updated_at`

	args := []any{
		project.ExternalID,
		project.ProposalID,
		project.Name,
		project.Status,
		geomText,
		pq.Array(project.Images),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&project.InternalID,
		&project.Version,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "project_project_id_key"`:
			return ErrDuplicateProjectID
		case err.Error() == `pq: duplicate key value violates unique constraint "project_proposal_id_key"`:
			return ErrDuplicateProposalID
		default:
			return err
		}
	}

	clientIDs := []int32{}
	for _, client := range project.Clients {
		clientIDs = append(clientIDs, *client.ClientID)
	}

	query = `
		INSERT INTO project_client (project_internal_id, client_internal_id)
		VALUES `

	args = []any{}
	for i, clientID := range clientIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, project.InternalID, clientID)
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
		return ErrZeroRowInserted
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m ProjectModel) Get(externalID int32) (*Project, error) {
	if externalID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT p.internal_id, p.project_id, p.proposal_id, p.name, p.status, 
		CASE 
			WHEN ST_IsEmpty(p.coordinates) THEN NULL 
			ELSE array[ST_X(p.coordinates), ST_Y(p.coordinates)] 
    	END AS coordinates, 
		array_agg(
			jsonb_build_object(
				'id', c.internal_id, 
				'name', c.name, 
				'address', c.address,
				'logo_url', c.logo_url,
				'note', c.note,
				'version', c.version,
				'created_at', c.created_at,
				'updated_at', c.updated_at
			)
		) as clients, p.images, p.version, p.created_at, p.updated_at 
		FROM project p
		LEFT JOIN project_client pc ON p.internal_id = pc.project_internal_id
		LEFT JOIN client c ON pc.client_internal_id = c.internal_id
		WHERE p.project_id = $1
		GROUP BY p.internal_id, p.project_id, p.proposal_id, p.name, p.status, p.coordinates, p.images, p.version, p.created_at, p.updated_at`

	var project Project
	var clients []string

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, externalID).Scan(
		&project.InternalID,
		&project.ExternalID,
		&project.ProposalID,
		&project.Name,
		&project.Status,
		pq.Array(&project.Coordinates),
		pq.Array(&clients),
		pq.Array(&project.Images),
		&project.Version,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	for _, client := range clients {
		var pc ProjectClient
		_ = json.Unmarshal([]byte(client), &pc)
		if pc.ClientID == nil {
			project.Clients = nil
			break
		}
		project.Clients = append(project.Clients, pc)
	}

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &project, nil
}

func (m ProjectModel) Update(project *Project) error {
	geomText := "POINT EMPTY"
	if project.Coordinates != nil {
		geomText = fmt.Sprintf("POINT(%f %f)", project.Coordinates[0], project.Coordinates[1])
	}
	if len(project.Images) == 0 {
		project.Images = nil
	}

	query := `
		UPDATE project
		SET project_id = $1, proposal_id = $2, name = $3, status = $4, coordinates = ST_GeomFromText($5, 4326), images = $6, version = version + 1, updated_at = $7
		WHERE internal_id = $8 AND version = $9
		RETURNING version, created_at, updated_at`

	args := []any{
		project.ExternalID,
		project.ProposalID,
		project.Name,
		project.Status,
		geomText,
		pq.Array(project.Images),
		time.Now(),
		project.InternalID,
		project.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&project.Version,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	query = `
		DELETE FROM project_client
		WHERE project_internal_id = $1 AND client_internal_id NOT IN (`
	args = []any{project.InternalID}
	for i, client := range project.Clients {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+2)
		args = append(args, client.ClientID)
	}
	query += ")"
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	query = `
		INSERT INTO project_client (project_internal_id, client_internal_id)
		VALUES `
	args = []any{}
	for i, client := range project.Clients {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, project.InternalID, client.ClientID)
	}
	query += `
		ON CONFLICT (project_internal_id, client_internal_id) DO NOTHING`
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

func (m ProjectModel) Delete(InternalID int32) error {
	query := `
		DELETE FROM project
		WHERE internal_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, InternalID)
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

type Coordinates [2]float64

type BoundingBox struct {
	BottomLeft Coordinates
	TopRight   Coordinates
	Valid      bool
}

func ConvertToBbox(bboxStrings []string) (BoundingBox, error) {
	bbox := BoundingBox{}

	// if bboxStrings == nil {
	// 	return bbox, nil
	// }
	// if len(bboxStrings) != 4 {
	// 	return bbox, ErrInvalidBBoxLength
	// }
	if bboxStrings != nil {
		for i, str := range bboxStrings {
			f, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return bbox, fmt.Errorf("error near comma %d: %w", i, err)
			}
			if i == 0 || i == 1 {
				bbox.BottomLeft[i] = f
			}
			if i == 2 || i == 3 {
				bbox.TopRight[i%2] = f
			}
		}
		bbox.Valid = true
	}
	return bbox, nil
}

func (m ProjectModel) GetAll(name string, status string, clients []string, bbox BoundingBox, filters Filters) ([]*Project, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), p.project_id, p.proposal_id, p.name, p.status,
		CASE
			WHEN ST_IsEmpty(p.coordinates) THEN NULL
			ELSE array[ST_X(p.coordinates), ST_Y(p.coordinates)]
		END AS coordinates,
		array_agg(
			jsonb_build_object(
				'id', c.internal_id,
				'name', c.name,
				'address', c.address,
				'logo_url', c.logo_url,
				'note', c.note,
				'version', c.version,
				'created_at', c.created_at,
				'updated_at', c.updated_at
			)
		) as clients, p.images, p.version, p.created_at, p.updated_at
		FROM project p
		LEFT JOIN project_client pc ON p.internal_id = pc.project_internal_id
		LEFT JOIN client c ON pc.client_internal_id = c.internal_id
		WHERE (
			( p.name ILIKE '%%' || $1 || '%%' OR $1 = '' )
			AND
			( to_tsvector('simple', p.status) @@ plainto_tsquery('simple', $2) or $2 = '' )
			AND
			( ($3::boolean IS FALSE)
				OR
				( 
					NOT ST_IsEmpty(p.coordinates) 
					AND ST_Within(p.coordinates, st_makeenvelope($4, $5, $6, $7, 4326)) 
				)
			)
		)
		GROUP BY p.internal_id, p.project_id, p.proposal_id, p.name, p.status, p.coordinates, p.images, p.version, p.created_at, p.updated_at
		ORDER BY %s %s, p.project_id ASC`,
		filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name,
		status,
		bbox.Valid,
		bbox.BottomLeft[0],
		bbox.BottomLeft[1],
		bbox.TopRight[0],
		bbox.TopRight[1],
	}

	if filters.limit() > 0 {
		query += `
			LIMIT $8 OFFSET $9`
		args = append(args, filters.limit(), filters.offset())
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	projects := []*Project{}

	for rows.Next() {
		var project Project
		var clients []string

		err := rows.Scan(
			&totalRecords,
			&project.ExternalID,
			&project.ProposalID,
			&project.Name,
			&project.Status,
			pq.Array(&project.Coordinates),
			pq.Array(&clients),
			pq.Array(&project.Images),
			&project.Version,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		for _, client := range clients {
			var s ProjectClient
			//TODO: handle unmarshal error
			_ = json.Unmarshal([]byte(client), &s)
			if s.ClientID == nil {
				project.Clients = nil
				break
			}
			project.Clients = append(project.Clients, s)
		}
		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return projects, metadata, nil
}
