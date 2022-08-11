package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/vietquy/alpha/errors"
	"github.com/vietquy/alpha/things"
)

var (
	// ErrSaveProject indicates error while saving to database
	ErrSaveProject = errors.New("save project to db error")
	// ErrUpdateProject indicates error while updating project in database
	ErrUpdateProject = errors.New("update project to db error")
	// ErrDeleteProject indicates error while deleting project in database
	ErrDeleteProject = errors.New("delete project from db error")
	// ErrSelectProject indicates error while reading project from database
	ErrSelectProject = errors.New("select project from db error")
	// ErrDeleteConnection indicates error while deleting connection in database
	ErrDeleteConnection = errors.New("unmarshal json error")
	// ErrHasThing indicates error while checking connection in database
	ErrHasThing = errors.New("check thing-project connection in database error")
	//ErrScan indicates error in database scanner
	ErrScan = errors.New("database scanner error")
	//ErrValue indicates error in database valuer
	ErrValue = errors.New("database valuer error")
)

var _ things.ProjectRepository = (*projectRepository)(nil)

type projectRepository struct {
	db *sqlx.DB
}

type dbConnection struct {
	Project string `db:"project"`
	Thing   string `db:"thing"`
	Owner   string `db:"owner"`
}

// NewProjectRepository instantiates a PostgreSQL implementation of project
// repository.
func NewProjectRepository(db *sqlx.DB) things.ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (cr projectRepository) Save(ctx context.Context, projects ...things.Project) ([]things.Project, error) {
	tx, err := cr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(ErrSaveProject, err)
	}

	q := `INSERT INTO projects (id, owner, name, metadata)
		  VALUES (:id, :owner, :name, :metadata);`

	for _, project := range projects {
		dbch := toDBProject(project)

		_, err = tx.NamedExecContext(ctx, q, dbch)
		if err != nil {
			tx.Rollback()
			pqErr, ok := err.(*pq.Error)
			if ok {
				switch pqErr.Code.Name() {
				case errInvalid, errTruncation:
					return []things.Project{}, things.ErrMalformedEntity
				case errDuplicate:
					return []things.Project{}, things.ErrConflict
				}
			}
			return []things.Project{}, errors.Wrap(ErrSaveProject, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return []things.Project{}, errors.Wrap(ErrSaveProject, err)
	}

	return projects, nil
}

func (cr projectRepository) Update(ctx context.Context, project things.Project) error {
	q := `UPDATE projects SET name = :name, metadata = :metadata WHERE owner = :owner AND id = :id;`

	dbch := toDBProject(project)

	res, err := cr.db.NamedExecContext(ctx, q, dbch)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return things.ErrMalformedEntity
			}
		}

		return errors.Wrap(ErrUpdateProject, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(ErrUpdateProject, err)
	}

	if cnt == 0 {
		return things.ErrNotFound
	}

	return nil
}

func (cr projectRepository) RetrieveByID(ctx context.Context, owner, id string) (things.Project, error) {
	q := `SELECT name, metadata FROM projects WHERE id = $1 AND owner = $2;`

	dbch := dbProject{
		ID:    id,
		Owner: owner,
	}
	if err := cr.db.QueryRowxContext(ctx, q, id, owner).StructScan(&dbch); err != nil {
		empty := things.Project{}
		pqErr, ok := err.(*pq.Error)
		if err == sql.ErrNoRows || ok && errInvalid == pqErr.Code.Name() {
			return empty, things.ErrNotFound
		}
		return empty, errors.Wrap(ErrSelectProject, err)
	}

	return toProject(dbch), nil
}

func (cr projectRepository) RetrieveAll(ctx context.Context, owner string, offset, limit uint64, name string, metadata things.Metadata) (things.ProjectsPage, error) {
	nq, name := getNameQuery(name)
	m, mq, err := getMetadataQuery(metadata)
	if err != nil {
		return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
	}

	q := fmt.Sprintf(`SELECT id, name, metadata FROM projects
	      WHERE owner = :owner %s%s ORDER BY id LIMIT :limit OFFSET :offset;`, mq, nq)

	params := map[string]interface{}{
		"owner":    owner,
		"limit":    limit,
		"offset":   offset,
		"name":     name,
		"metadata": m,
	}
	rows, err := cr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
	}
	defer rows.Close()

	items := []things.Project{}
	for rows.Next() {
		dbch := dbProject{Owner: owner}
		if err := rows.StructScan(&dbch); err != nil {
			return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
		}
		ch := toProject(dbch)

		items = append(items, ch)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM projects WHERE owner = :owner %s%s;`, nq, mq)

	total, err := total(ctx, cr.db, cq, params)
	if err != nil {
		return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
	}

	page := things.ProjectsPage{
		Projects: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return page, nil
}

func (cr projectRepository) RetrieveByThing(ctx context.Context, owner, thing string, offset, limit uint64) (things.ProjectsPage, error) {
	// Verify if UUID format is valid to avoid internal Postgres error
	if _, err := uuid.FromString(thing); err != nil {
		return things.ProjectsPage{}, things.ErrNotFound
	}

	q := `SELECT id, name, metadata
	      FROM projects ch
	      INNER JOIN connections co
		  ON ch.id = co.project_id
		  WHERE ch.owner = :owner AND co.thing_id = :thing
		  ORDER BY ch.id
		  LIMIT :limit
		  OFFSET :offset`

	params := map[string]interface{}{
		"owner":  owner,
		"thing":  thing,
		"limit":  limit,
		"offset": offset,
	}

	rows, err := cr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
	}
	defer rows.Close()

	items := []things.Project{}
	for rows.Next() {
		dbch := dbProject{Owner: owner}
		if err := rows.StructScan(&dbch); err != nil {
			return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
		}

		ch := toProject(dbch)
		items = append(items, ch)
	}

	q = `SELECT COUNT(*)
	     FROM projects ch
	     INNER JOIN connections co
	     ON ch.id = co.project_id
	     WHERE ch.owner = $1 AND co.thing_id = $2`

	var total uint64
	if err := cr.db.GetContext(ctx, &total, q, owner, thing); err != nil {
		return things.ProjectsPage{}, errors.Wrap(ErrSelectProject, err)
	}

	return things.ProjectsPage{
		Projects: items,
		PageMetadata: things.PageMetadata{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

func (cr projectRepository) Remove(ctx context.Context, owner, id string) error {
	dbch := dbProject{
		ID:    id,
		Owner: owner,
	}
	q := `DELETE FROM projects WHERE id = :id AND owner = :owner`
	cr.db.NamedExecContext(ctx, q, dbch)
	return nil
}

func (cr projectRepository) Connect(ctx context.Context, owner string, chIDs, thIDs []string) error {
	tx, err := cr.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(ErrDeleteProject, err)
	}

	q := `INSERT INTO connections (project_id, project_owner, thing_id, thing_owner)
	      VALUES (:project, :owner, :thing, :owner);`

	for _, chID := range chIDs {
		for _, thID := range thIDs {
			dbco := dbConnection{
				Project: chID,
				Thing:   thID,
				Owner:   owner,
			}

			_, err := tx.NamedExecContext(ctx, q, dbco)
			if err != nil {
				tx.Rollback()
				pqErr, ok := err.(*pq.Error)
				if ok {
					switch pqErr.Code.Name() {
					case errFK:
						return things.ErrNotFound
					case errDuplicate:
						return things.ErrConflict
					}
				}

				return errors.Wrap(ErrDeleteProject, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(ErrDeleteProject, err)
	}

	return nil
}

func (cr projectRepository) Disconnect(ctx context.Context, owner, projectID, thingID string) error {
	q := `DELETE FROM connections
	      WHERE project_id = :project AND project_owner = :owner
	      AND thing_id = :thing AND thing_owner = :owner`

	conn := dbConnection{
		Project: projectID,
		Thing:   thingID,
		Owner:   owner,
	}

	res, err := cr.db.NamedExecContext(ctx, q, conn)
	if err != nil {
		return errors.Wrap(ErrDeleteConnection, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(ErrDeleteConnection, err)
	}

	if cnt == 0 {
		return things.ErrNotFound
	}

	return nil
}

func (cr projectRepository) HasThing(ctx context.Context, projectID, key string) (string, error) {
	var thingID string
	q := `SELECT id FROM things WHERE key = $1`
	if err := cr.db.QueryRowxContext(ctx, q, key).Scan(&thingID); err != nil {
		return "", errors.Wrap(ErrHasThing, err)

	}

	if err := cr.hasThing(ctx, projectID, thingID); err != nil {
		return "", errors.Wrap(ErrHasThing, err)
	}

	return thingID, nil
}

func (cr projectRepository) HasThingByID(ctx context.Context, projectID, thingID string) error {
	return cr.hasThing(ctx, projectID, thingID)
}

func (cr projectRepository) hasThing(ctx context.Context, projectID, thingID string) error {
	q := `SELECT EXISTS (SELECT 1 FROM connections WHERE project_id = $1 AND thing_id = $2);`
	exists := false
	if err := cr.db.QueryRowxContext(ctx, q, projectID, thingID).Scan(&exists); err != nil {
		return errors.Wrap(ErrHasThing, err)
	}

	if !exists {
		return things.ErrUnauthorizedAccess
	}

	return nil
}

// dbMetadata type for handling metadata properly in database/sql.
type dbMetadata map[string]interface{}

// Scan implements the database/sql scanner interface.
func (m *dbMetadata) Scan(value interface{}) error {
	if value == nil {
		m = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		m = &dbMetadata{}
		return things.ErrScanMetadata
	}

	if err := json.Unmarshal(b, m); err != nil {
		m = &dbMetadata{}
		return err
	}

	return nil
}

// Value implements database/sql valuer interface.
func (m dbMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, err
}

type dbProject struct {
	ID       string     `db:"id"`
	Owner    string     `db:"owner"`
	Name     string     `db:"name"`
	Metadata dbMetadata `db:"metadata"`
}

func toDBProject(ch things.Project) dbProject {
	return dbProject{
		ID:       ch.ID,
		Owner:    ch.Owner,
		Name:     ch.Name,
		Metadata: ch.Metadata,
	}
}

func toProject(ch dbProject) things.Project {
	return things.Project{
		ID:       ch.ID,
		Owner:    ch.Owner,
		Name:     ch.Name,
		Metadata: ch.Metadata,
	}
}

func getNameQuery(name string) (string, string) {
	name = strings.ToLower(name)
	nq := ""
	if name != "" {
		name = fmt.Sprintf(`%%%s%%`, name)
		nq = ` AND LOWER(name) LIKE :name`
	}
	return nq, name
}

func getMetadataQuery(m things.Metadata) ([]byte, string, error) {
	mq := ""
	mb := []byte("{}")
	if len(m) > 0 {
		mq = ` AND metadata @> :metadata`

		b, err := json.Marshal(m)
		if err != nil {
			return nil, "", err
		}
		mb = b
	}
	return mb, mq, nil
}

func total(ctx context.Context, db *sqlx.DB, query string, params map[string]interface{}) (uint64, error) {
	rows, err := db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return 0, err
	}

	total := uint64(0)
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}

	return total, nil
}
