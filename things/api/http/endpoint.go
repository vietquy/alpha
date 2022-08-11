package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vietquy/alpha/things"
)

func createThingEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createThingReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		th := things.Thing{
			Key:      req.Key,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		saved, err := svc.CreateThings(ctx, req.token, th)
		if err != nil {
			return nil, err
		}

		res := thingRes{
			ID:      saved[0].ID,
			created: true,
		}

		return res, nil
	}
}

func createThingsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createThingsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		ths := []things.Thing{}
		for _, tReq := range req.Things {
			th := things.Thing{
				Name:     tReq.Name,
				Key:      tReq.Key,
				Metadata: tReq.Metadata,
			}
			ths = append(ths, th)
		}

		saved, err := svc.CreateThings(ctx, req.token, ths...)
		if err != nil {
			return nil, err
		}

		res := thingsRes{
			Things:  []thingRes{},
			created: true,
		}

		for _, th := range saved {
			tRes := thingRes{
				ID:       th.ID,
				Name:     th.Name,
				Key:      th.Key,
				Metadata: th.Metadata,
			}
			res.Things = append(res.Things, tRes)
		}

		return res, nil
	}
}

func updateThingEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateThingReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		thing := things.Thing{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}

		if err := svc.UpdateThing(ctx, req.token, thing); err != nil {
			return nil, err
		}

		res := thingRes{ID: req.id, created: false}
		return res, nil
	}
}

func updateKeyEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateKeyReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.UpdateKey(ctx, req.token, req.id, req.Key); err != nil {
			return nil, err
		}

		res := thingRes{ID: req.id, created: false}
		return res, nil
	}
}

func viewThingEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		thing, err := svc.ViewThing(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}

		res := viewThingRes{
			ID:       thing.ID,
			Owner:    thing.Owner,
			Name:     thing.Name,
			Key:      thing.Key,
			Metadata: thing.Metadata,
		}
		return res, nil
	}
}

func listThingsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listResourcesReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListThings(ctx, req.token, req.offset, req.limit, req.name, req.metadata)
		if err != nil {
			return nil, err
		}

		res := thingsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Things: []viewThingRes{},
		}
		for _, thing := range page.Things {
			view := viewThingRes{
				ID:       thing.ID,
				Owner:    thing.Owner,
				Name:     thing.Name,
				Key:      thing.Key,
				Metadata: thing.Metadata,
			}
			res.Things = append(res.Things, view)
		}

		return res, nil
	}
}

func listThingsByProjectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listByConnectionReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListThingsByProject(ctx, req.token, req.id, req.offset, req.limit)
		if err != nil {
			return nil, err
		}

		res := thingsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Things: []viewThingRes{},
		}
		for _, thing := range page.Things {
			view := viewThingRes{
				ID:       thing.ID,
				Owner:    thing.Owner,
				Key:      thing.Key,
				Name:     thing.Name,
				Metadata: thing.Metadata,
			}
			res.Things = append(res.Things, view)
		}

		return res, nil
	}
}

func removeThingEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		err := req.validate()
		if err == things.ErrNotFound {
			return removeRes{}, nil
		}

		if err != nil {
			return nil, err
		}

		if err := svc.RemoveThing(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return removeRes{}, nil
	}
}

func createProjectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createProjectReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		ch := things.Project{Name: req.Name, Metadata: req.Metadata}
		saved, err := svc.CreateProjects(ctx, req.token, ch)
		if err != nil {
			return nil, err
		}

		res := projectRes{
			ID:      saved[0].ID,
			created: true,
		}
		return res, nil
	}
}

func createProjectsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createProjectsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		chs := []things.Project{}
		for _, cReq := range req.Projects {
			ch := things.Project{
				Metadata: cReq.Metadata,
				Name:     cReq.Name,
			}
			chs = append(chs, ch)
		}

		saved, err := svc.CreateProjects(ctx, req.token, chs...)
		if err != nil {
			return nil, err
		}

		res := projectsRes{
			Projects: []projectRes{},
			created:  true,
		}

		for _, ch := range saved {
			cRes := projectRes{
				ID:       ch.ID,
				Name:     ch.Name,
				Metadata: ch.Metadata,
			}
			res.Projects = append(res.Projects, cRes)
		}

		return res, nil
	}
}

func updateProjectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateProjectReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		project := things.Project{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		if err := svc.UpdateProject(ctx, req.token, project); err != nil {
			return nil, err
		}

		res := projectRes{
			ID:      req.id,
			created: false,
		}
		return res, nil
	}
}

func viewProjectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		project, err := svc.ViewProject(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}

		res := viewProjectRes{
			ID:       project.ID,
			Owner:    project.Owner,
			Name:     project.Name,
			Metadata: project.Metadata,
		}

		return res, nil
	}
}

func listProjectsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listResourcesReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListProjects(ctx, req.token, req.offset, req.limit, req.name, req.metadata)
		if err != nil {
			return nil, err
		}

		res := projectsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Projects: []viewProjectRes{},
		}
		// Cast projects
		for _, project := range page.Projects {
			view := viewProjectRes{
				ID:       project.ID,
				Owner:    project.Owner,
				Name:     project.Name,
				Metadata: project.Metadata,
			}

			res.Projects = append(res.Projects, view)
		}

		return res, nil
	}
}

func listProjectsByThingEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listByConnectionReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListProjectsByThing(ctx, req.token, req.id, req.offset, req.limit)
		if err != nil {
			return nil, err
		}

		res := projectsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Projects: []viewProjectRes{},
		}
		for _, project := range page.Projects {
			view := viewProjectRes{
				ID:       project.ID,
				Owner:    project.Owner,
				Name:     project.Name,
				Metadata: project.Metadata,
			}
			res.Projects = append(res.Projects, view)
		}

		return res, nil
	}
}

func removeProjectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		if err := req.validate(); err != nil {
			if err == things.ErrNotFound {
				return removeRes{}, nil
			}
			return nil, err
		}

		if err := svc.RemoveProject(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return removeRes{}, nil
	}
}

func connectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cr := request.(connectionReq)

		if err := cr.validate(); err != nil {
			return nil, err
		}

		if err := svc.Connect(ctx, cr.token, []string{cr.projectID}, []string{cr.thingID}); err != nil {
			return nil, err
		}

		return connectionRes{}, nil
	}
}

func createConnectionsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cr := request.(createConnectionsReq)

		if err := cr.validate(); err != nil {
			return nil, err
		}

		if err := svc.Connect(ctx, cr.token, cr.ProjectIDs, cr.ThingIDs); err != nil {
			return nil, err
		}

		return createConnectionsRes{}, nil
	}
}

func disconnectEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cr := request.(connectionReq)

		if err := cr.validate(); err != nil {
			return nil, err
		}

		if err := svc.Disconnect(ctx, cr.token, cr.projectID, cr.thingID); err != nil {
			return nil, err
		}

		return disconnectionRes{}, nil
	}
}
