package things

import (
	"context"
)

// IdentityProvider specifies an API for generating unique identifiers.
type IdentityProvider interface {
	// ID generates the unique identifier.
	ID() (string, error)
}

// Metadata to be used for thing or project for customized
// describing of particular thing or project.
type Metadata map[string]interface{}

// Thing represents a thing. Each thing is owned by one user, and
// it is assigned with the unique identifier and (temporary) access key.
type Thing struct {
	ID       string
	Owner    string
	Name     string
	Key      string
	Metadata Metadata
}

// Page contains page related metadata as well as list of things that
// belong to this page.
type Page struct {
	PageMetadata
	Things []Thing
}

// Project represents a "communication group". This group contains the
// things that can exchange messages between eachother.
type Project struct {
	ID       string
	Owner    string
	Name     string
	Metadata map[string]interface{}
}

// ProjectsPage contains page related metadata as well as list of projects that
// belong to this page.
type ProjectsPage struct {
	PageMetadata
	Projects []Project
}

// ThingRepository specifies a thing persistence API.
type ThingRepository interface {
	// Save persists multiple things. Things are saved using a transaction. If one thing
	// fails then none will be saved. Successful operation is indicated by non-nil
	// error response.
	Save(ctx context.Context, ths ...Thing) ([]Thing, error)

	// Update performs an update to the existing thing. A non-nil error is
	// returned to indicate operation failure.
	Update(ctx context.Context, t Thing) error

	// UpdateKey updates key value of the existing thing. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(ctx context.Context, owner, id, key string) error

	// RetrieveByID retrieves the thing having the provided identifier, that is owned
	// by the specified user.
	RetrieveByID(ctx context.Context, owner, id string) (Thing, error)

	// RetrieveByKey returns thing ID for given thing key.
	RetrieveByKey(ctx context.Context, key string) (string, error)

	// RetrieveAll retrieves the subset of things owned by the specified user.
	RetrieveAll(ctx context.Context, owner string, offset, limit uint64, name string, m Metadata) (Page, error)

	// RetrieveByProject retrieves the subset of things owned by the specified
	// user and connected to specified project.
	RetrieveByProject(ctx context.Context, owner, project string, offset, limit uint64) (Page, error)

	// Remove removes the thing having the provided identifier, that is owned
	// by the specified user.
	Remove(ctx context.Context, owner, id string) error
}

// ProjectRepository specifies a project persistence API.
type ProjectRepository interface {
	// Save persists multiple projects. Projects are saved using a transaction. If one project
	// fails then none will be saved. Successful operation is indicated by non-nil
	// error response.
	Save(context.Context, ...Project) ([]Project, error)

	// Update performs an update to the existing project. A non-nil error is
	// returned to indicate operation failure.
	Update(context.Context, Project) error

	// RetrieveByID retrieves the project having the provided identifier, that is owned
	// by the specified user.
	RetrieveByID(context.Context, string, string) (Project, error)

	// RetrieveAll retrieves the subset of projects owned by the specified user.
	RetrieveAll(context.Context, string, uint64, uint64, string, Metadata) (ProjectsPage, error)

	// RetrieveByThing retrieves the subset of projects owned by the specified
	// user and have specified thing connected to them.
	RetrieveByThing(context.Context, string, string, uint64, uint64) (ProjectsPage, error)

	// Remove removes the project having the provided identifier, that is owned
	// by the specified user.
	Remove(context.Context, string, string) error

	// Connect adds things to the project's list of connected things.
	Connect(context.Context, string, []string, []string) error

	// Disconnect removes thing from the project's list of connected
	// things.
	Disconnect(context.Context, string, string, string) error

	// HasThing determines whether the thing with the provided access key, is
	// "connected" to the specified project. If that's the case, it returns
	// thing's ID.
	HasThing(context.Context, string, string) (string, error)

	// HasThingByID determines whether the thing with the provided ID, is
	// "connected" to the specified project. If that's the case, then
	// returned error will be nil.
	HasThingByID(context.Context, string, string) error
}
