package http

import (
	"github.com/vietquy/alpha/things"
)

const maxLimitSize = 100
const maxNameSize = 1024

type apiReq interface {
	validate() error
}

type createThingReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req createThingReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type createThingsReq struct {
	token  string
	Things []createThingReq
}

func (req createThingsReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if len(req.Things) <= 0 {
		return things.ErrMalformedEntity
	}

	for _, thing := range req.Things {
		if len(thing.Name) > maxNameSize {
			return things.ErrMalformedEntity
		}
	}

	return nil
}

type updateThingReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateThingReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type updateKeyReq struct {
	token string
	id    string
	Key   string `json:"key"`
}

func (req updateKeyReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" || req.Key == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type createProjectReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req createProjectReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type createProjectsReq struct {
	token    string
	Projects []createProjectReq
}

func (req createProjectsReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if len(req.Projects) <= 0 {
		return things.ErrMalformedEntity
	}

	for _, project := range req.Projects {
		if len(project.Name) > maxNameSize {
			return things.ErrMalformedEntity
		}
	}

	return nil
}

type updateProjectReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateProjectReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type viewResourceReq struct {
	token string
	id    string
}

func (req viewResourceReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type listResourcesReq struct {
	token    string
	offset   uint64
	limit    uint64
	name     string
	metadata map[string]interface{}
}

func (req *listResourcesReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.limit == 0 || req.limit > maxLimitSize {
		return things.ErrMalformedEntity
	}

	if len(req.name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type listByConnectionReq struct {
	token  string
	id     string
	offset uint64
	limit  uint64
}

func (req listByConnectionReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	if req.limit == 0 || req.limit > maxLimitSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type connectionReq struct {
	token   string
	projectID  string
	thingID string
}

func (req connectionReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.projectID == "" || req.thingID == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type createConnectionsReq struct {
	token      string
	ProjectIDs []string `json:"project_ids,omitempty"`
	ThingIDs   []string `json:"thing_ids,omitempty"`
}

func (req createConnectionsReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if len(req.ProjectIDs) == 0 || len(req.ThingIDs) == 0 {
		return things.ErrMalformedEntity
	}

	for _, chID := range req.ProjectIDs {
		if chID == "" {
			return things.ErrMalformedEntity
		}
	}
	for _, thingID := range req.ThingIDs {
		if thingID == "" {
			return things.ErrMalformedEntity
		}
	}

	return nil
}
