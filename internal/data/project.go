package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hwanbin/wanpm-api/internal/s3action"
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
	ExternalID  *int32                     `json:"project_id"`
	ProposalID  *string                    `json:"proposal_id"`
	Name        *string                    `json:"name"`
	Status      *string                    `json:"status"`
	Feature     *Feature                   `json:"feature"`
	Note        *string                    `json:"note"`
	Images      []string                   `json:"images"`
	ClientNames []string                   `json:"client_names"`
	Assignments []ProjectAssignmentRequest `json:"assignments"`
}

type ProjectAssignmentRequest struct {
	EmployeeID *string `json:"employee_id"`
	RoleID     *int32  `json:"role_id"`
}

type ProjectAssignment struct {
	EmployeeID    string `json:"employee_id"`
	EmployeeEmail string `json:"employee_email"`
	RoleID        int32  `json:"role_id"`
	RoleName      string `json:"role_name"`
}

type ProjectRequest struct {
	InternalID  int32                      `json:"-"`
	ExternalID  *int32                     `json:"project_id"`
	ProposalID  *string                    `json:"proposal_id"`
	Name        *string                    `json:"name"`
	Status      *string                    `json:"status"`
	Feature     *string                    `json:"feature"`
	Note        *string                    `json:"note"`
	Images      []string                   `json:"images"`
	Clients     []ProjectClient            `json:"clients"`
	Assignments []ProjectAssignmentRequest `json:"assignments"`
	Version     int32                      `json:"version"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

type ProjectResponse struct {
	InternalID  int32               `json:"-"`
	ExternalID  *int32              `json:"project_id"`
	ProposalID  *string             `json:"proposal_id"`
	Name        *string             `json:"name"`
	Status      *string             `json:"status"`
	Feature     *Feature            `json:"feature"`
	Note        *string             `json:"note"`
	Images      []string            `json:"images"`
	Clients     []ProjectClient     `json:"clients"`
	Assignments []ProjectAssignment `json:"assignments"`
	Version     int32               `json:"version"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type ProjectQsInput struct {
	Name        string
	Status      string
	ProposalId  string
	ProjectId   string
	FullAddress string
	ClientName  string
	Clients     []string
	Bbox        []string
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
	if p.Feature != nil {
		v.Check(p.Feature.IsValidFeature(), "feature", "must be valid feature json")

		if !p.Feature.IsEmptyFeature() {
			v.Check(p.Feature.IsValidFeatureType(), "feat_type", "must be 'Feature' type")
			v.Check(p.Feature.IsValidPointGeometryType(), "geom_type", "must be 'Point' type")
			v.Check(p.Feature.IsValidPointGeometryCoordinates(), "geom_coords", "must be a lng, lat pair")
			if len(p.Feature.Geometry.Coordinates) == 2 {
				v.Check(p.Feature.IsValidLongitude(), "geom_coords_lng", "longitude must be in between -180 and 180")
				v.Check(p.Feature.IsValidLatitude(), "geom_coords_lat", "latitude must be in between -90 and 90")
			}
			v.Check(p.Feature.IsValidPropertiesName(), "props_name", "invalid or missing 'name' field")
			v.Check(p.Feature.IsValidPropertiesFullAddress(), "props_full_addr", "invalid or missing 'full_address' field")
		}
	}

	v.Check(*p.ExternalID > 0, "project_id", "must be a positive integer")

	v.Check(*p.ProposalID != "", "proposal_id", "must not be an empty string")
	v.Check(len(*p.ProposalID) <= 10, "proposal_id", "must not be more than 10 bytes long")

	v.Check(*p.Name != "", "name", "must not be an empty string")
	v.Check(len(*p.Name) <= 500, "name", "must not be more than 500 bytes long")

	v.Check(*p.Status != "", "status", "must not be an empty string")
	v.Check(len(*p.Status) <= 100, "status", "must not be more than 100 bytes long")

	if p.Note != nil {
		v.Check(len(*p.Note) <= 1000, "note", "must not be more than 1000 bytes long")
	}

	v.Check(len(p.ClientNames) >= 1, "client_names", "must contain at least 1 client")
	v.Check(validator.Unique(p.ClientNames), "client_names", "must not contain duplicate values")

	if p.Images != nil {
		v.Check(len(p.Images) >= 1, "images", "must contain at least 1 image")
		v.Check(validator.Unique(p.Images), "images", "must not contain duplicate values")
	}

	if len(p.Assignments) > 0 {
		for _, assignment := range p.Assignments {
			v.Check(assignment.EmployeeID != nil, "assignment employee_id", "must be provided")
			v.Check(assignment.RoleID != nil, "assignment role_id", "must be provided")
		}
	}
}

func ValidateUpdatingProjectInput(v *validator.Validator, p *ProjectInput) {
	if p.ExternalID != nil {
		v.Check(*p.ExternalID > 0, "project_id", "must be a positive integer")
	}

	if p.ProposalID != nil {
		v.Check(*p.ProposalID != "", "proposal_id", "must not be empty string")
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

	if p.Feature != nil {
		v.Check(p.Feature.IsValidFeature(), "feature", "must be valid feature json")

		if !p.Feature.IsEmptyFeature() {
			v.Check(p.Feature.IsValidFeatureType(), "feat_type", "must be 'Feature' type")
			v.Check(p.Feature.IsValidPointGeometryType(), "geom_type", "must be 'Point' type")
			v.Check(p.Feature.IsValidPointGeometryCoordinates(), "geom_coords", "must be a lng, lat pair")
			if len(p.Feature.Geometry.Coordinates) == 2 {
				v.Check(p.Feature.IsValidLongitude(), "geom_coords_lng", "longitude must be in between -180 and 180")
				v.Check(p.Feature.IsValidLatitude(), "geom_coords_lat", "latitude must be in between -90 and 90")
			}
			v.Check(p.Feature.IsValidPropertiesName(), "props_name", "invalid or missing 'name' field")
			v.Check(p.Feature.IsValidPropertiesFullAddress(), "props_full_addr", "invalid or missing 'full_address' field")
		} else {
			p.Feature = nil
		}
	}

	if len(p.Images) > 0 {
		v.Check(validator.Unique(p.Images), "images", "must not contain duplicate values")
	}
}

func ValidateProject(v *validator.Validator, p *ProjectRequest) {
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

func (m ProjectModel) Insert(project *ProjectRequest) error {
	query := `
		INSERT INTO project (project_id, proposal_id, name, status, feature, images, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING internal_id, version, created_at, updated_at`
	args := []any{
		project.ExternalID,
		project.ProposalID,
		project.Name,
		project.Status,
		project.Feature,
		pq.Array(project.Images),
		project.Note,
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

	// inserting clients
	clientIDs := []int32{}
	for _, client := range project.Clients {
		clientIDs = append(clientIDs, *client.ClientID)
	}

	query = `
		INSERT INTO project_client (project_internal_id, client_id)
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
	// end of inserting clients

	// inserting assignments
	if len(project.Assignments) > 0 {
		query = `
		INSERT INTO assignment (project_id, employee_id, role_id)
		VALUES `
		args = []any{}
		for i, assignment := range project.Assignments {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
			args = append(args, project.ExternalID, *assignment.EmployeeID, *assignment.RoleID)
		}

		result, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil
		}

		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return ErrZeroRowInserted
		}
	}
	// end of inserting assignments

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m ProjectModel) Get(externalID int32) (*ProjectResponse, error) {
	if externalID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT p.internal_id, p.project_id, p.proposal_id, p.name, p.status, p.feature, p.note,
		COALESCE(pc.clients, ARRAY[]::jsonb[]) AS clients, 
		COALESCE(pa.assignments, ARRAY[]::jsonb[]) AS assignments,
		p.images, p.version, p.created_at, p.updated_at 
		FROM project p
		LEFT JOIN (
			SELECT 
				pc.project_internal_id,
				array_agg(
				 	jsonb_build_object(
						'id', c.id, 
						'name', c.name, 
						'address', c.address,
						'logo_url', c.logo_url,
						'note', c.note,
						'version', c.version,
						'created_at', c.created_at,
						'updated_at', c.updated_at
					)
				) AS clients
			FROM project_client pc
			JOIN client c ON pc.client_id = c.id
			GROUP BY pc.project_internal_id
		) pc ON p.internal_id = pc.project_internal_id
		LEFT JOIN (
			SELECT 
				a.project_id,
				array_agg(
					jsonb_build_object(
						'employee_id', u.id,
						'employee_email', u.email,
						'role_id', r.id,
						'role_name', r.name
					)
				) AS assignments
			FROM assignment a
			JOIN appuser u ON a.employee_id = u.id
			JOIN role r ON a.role_id = r.id
			GROUP BY a.project_id
		) pa ON p.project_id = pa.project_id		
		WHERE p.project_id = $1`

	var project ProjectResponse
	var projectFeature string
	var clients []string
	var assignments []string

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, externalID).Scan(
		&project.InternalID,
		&project.ExternalID,
		&project.ProposalID,
		&project.Name,
		&project.Status,
		&projectFeature,
		&project.Note,
		pq.Array(&clients),
		pq.Array(&assignments),
		pq.Array(&project.Images),
		&project.Version,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	err = json.Unmarshal([]byte(projectFeature), &project.Feature)
	if err != nil {
		return nil, errors.New("failed to unmarshal feature")
	}

	for _, client := range clients {
		var pc ProjectClient
		err = json.Unmarshal([]byte(client), &pc)
		if err != nil {
			return nil, errors.New("failed to unmarshal clients associated with the project")
		}
		if pc.ClientID == nil {
			project.Clients = nil
			break
		}
		project.Clients = append(project.Clients, pc)
	}

	for _, assignment := range assignments {
		var pa ProjectAssignment
		err = json.Unmarshal([]byte(assignment), &pa)
		if err != nil {
			return nil, errors.New("failed to unmarshal assignments associated with the project")
		}
		project.Assignments = append(project.Assignments, pa)
	}

	return &project, nil
}

func (m ProjectModel) Update(project *ProjectRequest) error {
	query := `
		UPDATE project
		SET project_id = $1, proposal_id = $2, name = $3, status = $4, feature = $5, images = $6, version = version + 1, updated_at = $7
		WHERE internal_id = $8 AND version = $9
		RETURNING version, created_at, updated_at`

	args := []any{
		project.ExternalID,
		project.ProposalID,
		project.Name,
		project.Status,
		project.Feature,
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

	// start of updating project_client
	query = `
		DELETE FROM project_client
		WHERE project_internal_id = $1 AND client_id NOT IN (`
	args = []any{project.InternalID}
	for i, client := range project.Clients {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+2)
		args = append(args, *client.ClientID)
	}
	query += ")"

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	query = `
		INSERT INTO project_client (project_internal_id, client_id)
		VALUES `
	args = []any{}
	for i, client := range project.Clients {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, project.InternalID, *client.ClientID)
	}
	query += `
		ON CONFLICT (project_internal_id, client_id) DO NOTHING`

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	// end of updating project_client

	// start of updating assignments
	if project.Assignments != nil {
		query = `
			DELETE FROM assignment
			WHERE project_id = $1`
		args = []any{project.ExternalID}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}

		if len(project.Assignments) != 0 {
			query = `
			INSERT INTO assignment (project_id, employee_id, role_id)
			VALUES `
			args = []any{}
			for i, assignment := range project.Assignments {
				if i > 0 {
					query += ", "
				}
				query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
				args = append(args, project.ExternalID, assignment.EmployeeID, assignment.RoleID)
			}
			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return err
			}
		}
	}
	// end of updating assignments

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m ProjectModel) Delete(InternalID int32, bucket, prefix string, client *s3.Client, objects []types.ObjectIdentifier) error {
	query := `
		DELETE FROM project
		WHERE internal_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// set type of all prefixed objects as `delete marker` by delete objects
	err = s3action.DeleteObjects(ctx, client, bucket, objects)
	if err != nil {
		restoringErr := s3action.RestoreDeletedObjects(ctx, client, bucket, prefix)
		if restoringErr != nil {
			return restoringErr
		}
		return err
	}

	result, err := tx.ExecContext(ctx, query, InternalID)
	if err != nil {
		restoringErr := s3action.RestoreDeletedObjects(ctx, client, bucket, prefix)
		if restoringErr != nil {
			return restoringErr
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		restoringErr := s3action.RestoreDeletedObjects(ctx, client, bucket, prefix)
		if restoringErr != nil {
			return restoringErr
		}
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	s3action.PermanentlyDeleteObjects(ctx, client, bucket, prefix)

	err = tx.Commit()
	if err != nil {
		return err
	}

	return err
}

type Coordinates [2]float64

type BoundingBox struct {
	BottomLeft Coordinates
	TopRight   Coordinates
	Valid      bool
}

func ConvertToBbox(bboxStrings []string) (BoundingBox, error) {
	bbox := BoundingBox{}

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

func (m ProjectModel) GetAll(qs ProjectQsInput, bbox BoundingBox) ([]*ProjectResponse, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), 
			p.project_id, p.proposal_id, p.name, p.status, p.feature,
			COALESCE(pc.clients, ARRAY[]::jsonb[]) AS clients, 
			COALESCE(pa.assignments, ARRAY[]::jsonb[]) AS assignments,
			p.images, p.version, p.created_at, p.updated_at
		FROM project p
		LEFT JOIN (
			SELECT 
				pc.project_internal_id,
				array_agg(
				 	jsonb_build_object(
						'id', c.id, 
						'name', c.name, 
						'address', c.address,
						'logo_url', c.logo_url,
						'note', c.note,
						'version', c.version,
						'created_at', c.created_at,
						'updated_at', c.updated_at
					)
				) AS clients
			FROM project_client pc
			JOIN client c ON pc.client_id = c.id
			GROUP BY pc.project_internal_id
		) pc ON p.internal_id = pc.project_internal_id
		LEFT JOIN (
			SELECT 
				a.project_id,
				array_agg(
					jsonb_build_object(
						'employee_id', u.id,
						'employee_email', u.email,
						'role_id', r.id,
						'role_name', r.name
					)
				) AS assignments
			FROM assignment a
			JOIN appuser u ON a.employee_id = u.id
			JOIN role r ON a.role_id = r.id
			GROUP BY a.project_id
		) pa ON p.project_id = pa.project_id
		WHERE (
			(
				( p.name ILIKE '%%' || $1 || '%%' and not $1 = '' )
				OR
				( p.status ILIKE '%%' || $2 || '%%' and not $2 = '' )
				OR
				( 
					($3::boolean IS TRUE)
					AND
					(
						jsonb_typeof(p.feature->'geometry'->'coordinates') = 'array'
						AND jsonb_array_length(p.feature->'geometry'->'coordinates') = 2
						AND ST_Within(
							ST_GeomFromGeoJSON(p.feature->'geometry'), 
							ST_MakeEnvelope($4, $5, $6, $7, 4326)
						)
					)
				)
				OR
				( CAST(p.project_id AS TEXT) LIKE '%%' || $8 || '%%' and not $8 = '' )
				OR
				( p.proposal_id ILIKE '%%' || $9 || '%%' and not $9 = '' )
				OR
				( p.feature->'properties'->>'full_address' ILIKE '%%' || $10 || '%%' and not $10 = '' )
				OR
				( 
					p.internal_id IN (
						SELECT pc.project_internal_id
						FROM project_client pc
						WHERE pc.client_id IN (
							SELECT c.id
							FROM client c
							WHERE ( c.name ILIKE '%%' || $11 || '%%' and not $11 = '' )
						)
					)
				)
			)
			OR
			( $1 = '' and $2 = '' and $3 = FALSE and $8 = '' and $9 = '' and $10 = '' and $11 = '' )
		)
		ORDER BY %s %s, p.project_id ASC`,
		qs.Filters.sortColumn(), qs.Filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		qs.Name,
		qs.Status,
		bbox.Valid,
		bbox.BottomLeft[0],
		bbox.BottomLeft[1],
		bbox.TopRight[0],
		bbox.TopRight[1],
		qs.ProjectId,
		qs.ProposalId,
		qs.FullAddress,
		qs.ClientName,
	}

	if qs.Filters.limit() > 0 {
		query += `
			LIMIT $12 OFFSET $13`
		args = append(args, qs.Filters.limit(), qs.Filters.offset())
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	projects := []*ProjectResponse{}

	for rows.Next() {
		var project ProjectResponse
		var projectFeature string
		var clients []string
		var assignments []string

		err := rows.Scan(
			&totalRecords,
			&project.ExternalID,
			&project.ProposalID,
			&project.Name,
			&project.Status,
			&projectFeature,
			pq.Array(&clients),
			pq.Array(&assignments),
			pq.Array(&project.Images),
			&project.Version,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		err = json.Unmarshal([]byte(projectFeature), &project.Feature)
		if err != nil {
			return nil, Metadata{}, errors.New("failed to unmarshal feature")
		}

		for _, client := range clients {
			var pc ProjectClient
			err = json.Unmarshal([]byte(client), &pc)
			if err != nil {
				return nil, Metadata{}, errors.New("failed to unmarshal clients")
			}
			if pc.ClientID == nil {
				project.Clients = nil
				break
			}
			project.Clients = append(project.Clients, pc)
		}

		for _, assignment := range assignments {
			var pa ProjectAssignment
			err = json.Unmarshal([]byte(assignment), &pa)
			if err != nil {
				return nil, Metadata{}, errors.New("failed to unmarshal assignments associated with the project")
			}
			project.Assignments = append(project.Assignments, pa)
		}

		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, qs.Filters.Page, qs.Filters.PageSize)

	return projects, metadata, nil
}
