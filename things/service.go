package things

import (
	"context"

	"github.com/vietquy/alpha/errors"

	"github.com/vietquy/alpha"
)

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")

	// ErrScanMetadata indicates problem with metadata in db
	ErrScanMetadata = errors.New("failed to scan metadata")

	// ErrCreateThings indicates error in creating Thing
	ErrCreateThings = errors.New("create thing failed")

	// ErrCreateProjects indicates error in creating Project
	ErrCreateProjects = errors.New("create project failed")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThings adds a list of things to the user identified by the provided key.
	CreateThings(ctx context.Context, token string, things ...Thing) ([]Thing, error)

	// UpdateThing updates the thing identified by the provided ID, that
	// belongs to the user identified by the provided key.
	UpdateThing(ctx context.Context, token string, thing Thing) error

	// UpdateKey updates key value of the existing thing. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(ctx context.Context, token, id, key string) error

	// ViewThing retrieves data about the thing identified with the provided
	// ID, that belongs to the user identified by the provided key.
	ViewThing(ctx context.Context, token, id string) (Thing, error)

	// ListThings retrieves data about subset of things that belongs to the
	// user identified by the provided key.
	ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata Metadata) (Page, error)

	// ListThingsByProject retrieves data about subset of things that are
	// connected to specified project and belong to the user identified by
	// the provided key.
	ListThingsByProject(ctx context.Context, token, project string, offset, limit uint64) (Page, error)

	// RemoveThing removes the thing identified with the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveThing(ctx context.Context, token, id string) error

	// CreateProjects adds a list of projects to the user identified by the provided key.
	CreateProjects(ctx context.Context, token string, projects ...Project) ([]Project, error)

	// UpdateProject updates the project identified by the provided ID, that
	// belongs to the user identified by the provided key.
	UpdateProject(ctx context.Context, token string, project Project) error

	// ViewProject retrieves data about the project identified by the provided
	// ID, that belongs to the user identified by the provided key.
	ViewProject(ctx context.Context, token, id string) (Project, error)

	// ListProjects retrieves data about subset of projects that belongs to the
	// user identified by the provided key.
	ListProjects(ctx context.Context, token string, offset, limit uint64, name string, m Metadata) (ProjectsPage, error)

	// ListProjectsByThing retrieves data about subset of projects that have
	// specified thing connected to them and belong to the user identified by
	// the provided key.
	ListProjectsByThing(ctx context.Context, token, thing string, offset, limit uint64) (ProjectsPage, error)

	// RemoveProject removes the thing identified by the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveProject(ctx context.Context, token, id string) error

	// Connect adds things to the project's list of connected things.
	Connect(ctx context.Context, token string, chIDs, thIDs []string) error

	// Disconnect removes thing from the project's list of connected
	// things.
	Disconnect(ctx context.Context, token, projectID, thingID string) error

	// CanAccessByKey determines whether the project can be accessed using the
	// provided key and returns thing's id if access is allowed.
	CanAccessByKey(ctx context.Context, projectID, key string) (string, error)

	// CanAccessByID determines whether the project can be accessed by
	// the given thing and returns error if it cannot.
	CanAccessByID(ctx context.Context, projectID, thingID string) error

	// Identify returns thing ID for given thing key.
	Identify(ctx context.Context, key string) (string, error)
}

// PageMetadata contains page metadata that helps navigation.
type PageMetadata struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Name   string
}

var _ Service = (*thingsService)(nil)

type thingsService struct {
	auth         alpha.AuthNServiceClient
	things       ThingRepository
	projects     ProjectRepository
	idp          IdentityProvider
}

// New instantiates the things service implementation.
func New(auth alpha.AuthNServiceClient, things ThingRepository, projects ProjectRepository, idp IdentityProvider) Service {
	return &thingsService{
		auth:         auth,
		things:       things,
		projects:     projects,
		idp:          idp,
	}
}

func (ts *thingsService) CreateThings(ctx context.Context, token string, things ...Thing) ([]Thing, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return []Thing{}, ErrUnauthorizedAccess
	}

	for i := range things {
		things[i].ID, err = ts.idp.ID()
		if err != nil {
			return []Thing{}, errors.Wrap(ErrCreateThings, err)
		}

		things[i].Owner = res.GetValue()

		if things[i].Key == "" {
			things[i].Key, err = ts.idp.ID()
			if err != nil {
				return []Thing{}, errors.Wrap(ErrCreateThings, err)
			}
		}
	}

	return ts.things.Save(ctx, things...)
}

func (ts *thingsService) UpdateThing(ctx context.Context, token string, thing Thing) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	thing.Owner = res.GetValue()

	return ts.things.Update(ctx, thing)
}

func (ts *thingsService) UpdateKey(ctx context.Context, token, id, key string) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	owner := res.GetValue()

	return ts.things.UpdateKey(ctx, owner, id, key)

}

func (ts *thingsService) ViewThing(ctx context.Context, token, id string) (Thing, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return Thing{}, ErrUnauthorizedAccess
	}

	return ts.things.RetrieveByID(ctx, res.GetValue(), id)
}

func (ts *thingsService) ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata Metadata) (Page, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return Page{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	// tp, err := ts.things.RetrieveAll(ctx, res.GetValue(), offset, limit, name, metadata)
	// return tp, errors.Wrap(ErrUnauthorizedAccess, err)
	return ts.things.RetrieveAll(ctx, res.GetValue(), offset, limit, name, metadata)
}

func (ts *thingsService) ListThingsByProject(ctx context.Context, token, project string, offset, limit uint64) (Page, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return Page{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return ts.things.RetrieveByProject(ctx, res.GetValue(), project, offset, limit)
}

func (ts *thingsService) RemoveThing(ctx context.Context, token, id string) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return ts.things.Remove(ctx, res.GetValue(), id)
}

func (ts *thingsService) CreateProjects(ctx context.Context, token string, projects ...Project) ([]Project, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return []Project{}, ErrUnauthorizedAccess
	}

	for i := range projects {
		projects[i].ID, err = ts.idp.ID()
		if err != nil {
			return []Project{}, errors.Wrap(ErrCreateProjects, err)
		}

		projects[i].Owner = res.GetValue()
	}

	return ts.projects.Save(ctx, projects...)
}

func (ts *thingsService) UpdateProject(ctx context.Context, token string, project Project) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	project.Owner = res.GetValue()
	return ts.projects.Update(ctx, project)
}

func (ts *thingsService) ViewProject(ctx context.Context, token, id string) (Project, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return Project{}, ErrUnauthorizedAccess
	}

	return ts.projects.RetrieveByID(ctx, res.GetValue(), id)
}

func (ts *thingsService) ListProjects(ctx context.Context, token string, offset, limit uint64, name string, m Metadata) (ProjectsPage, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ProjectsPage{}, ErrUnauthorizedAccess
	}

	return ts.projects.RetrieveAll(ctx, res.GetValue(), offset, limit, name, m)
}

func (ts *thingsService) ListProjectsByThing(ctx context.Context, token, thing string, offset, limit uint64) (ProjectsPage, error) {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ProjectsPage{}, ErrUnauthorizedAccess
	}

	return ts.projects.RetrieveByThing(ctx, res.GetValue(), thing, offset, limit)
}

func (ts *thingsService) RemoveProject(ctx context.Context, token, id string) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	return ts.projects.Remove(ctx, res.GetValue(), id)
}

func (ts *thingsService) Connect(ctx context.Context, token string, chIDs, thIDs []string) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	return ts.projects.Connect(ctx, res.GetValue(), chIDs, thIDs)
}

func (ts *thingsService) Disconnect(ctx context.Context, token, projectID, thingID string) error {
	res, err := ts.auth.Identify(ctx, &alpha.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	return ts.projects.Disconnect(ctx, res.GetValue(), projectID, thingID)
}

func (ts *thingsService) CanAccessByKey(ctx context.Context, projectID, key string) (string, error) {
	thingID, err := ts.projects.HasThing(ctx, projectID, key)
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	return thingID, nil
}

func (ts *thingsService) CanAccessByID(ctx context.Context, projectID, thingID string) error {
	if err := ts.projects.HasThingByID(ctx, projectID, thingID); err != nil {
		return ErrUnauthorizedAccess
	}

	return nil
}

func (ts *thingsService) Identify(ctx context.Context, key string) (string, error) {
	id, err := ts.things.RetrieveByKey(ctx, key)
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	return id, nil
}
