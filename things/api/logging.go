package api

import (
	"context"
	"fmt"
	"time"

	log "github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/things"
)

var _ things.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    things.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc things.Service, logger log.Logger) things.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) CreateThings(ctx context.Context, token string, ths ...things.Thing) (saved []things.Thing, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_things for token %s and things %s took %s to complete", token, saved, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateThings(ctx, token, ths...)
}

func (lm *loggingMiddleware) UpdateThing(ctx context.Context, token string, thing things.Thing) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_thing for token %s and thing %s took %s to complete", token, thing.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateThing(ctx, token, thing)
}

func (lm *loggingMiddleware) UpdateKey(ctx context.Context, token, id, key string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_key for thing %s and key %s took %s to complete", id, key, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateKey(ctx, token, id, key)
}

func (lm *loggingMiddleware) ViewThing(ctx context.Context, token, id string) (thing things.Thing, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_thing for token %s and thing %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewThing(ctx, token, id)
}

func (lm *loggingMiddleware) ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata things.Metadata) (_ things.Page, err error) {
	defer func(begin time.Time) {
		nlog := ""
		if name != "" {
			nlog = fmt.Sprintf("with name %s ", name)
		}
		message := fmt.Sprintf("Method list_things %sfor token %s took %s to complete", nlog, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListThings(ctx, token, offset, limit, name, metadata)
}

func (lm *loggingMiddleware) ListThingsByProject(ctx context.Context, token, id string, offset, limit uint64) (_ things.Page, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_things_by_project for project %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s", message, err))
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListThingsByProject(ctx, token, id, offset, limit)
}

func (lm *loggingMiddleware) RemoveThing(ctx context.Context, token, id string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove_thing for token %s and thing %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveThing(ctx, token, id)
}

func (lm *loggingMiddleware) CreateProjects(ctx context.Context, token string, projects ...things.Project) (saved []things.Project, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_projects for token %s and projects %s took %s to complete", token, saved, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateProjects(ctx, token, projects...)
}

func (lm *loggingMiddleware) UpdateProject(ctx context.Context, token string, project things.Project) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_project for token %s and project %s took %s to complete", token, project.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateProject(ctx, token, project)
}

func (lm *loggingMiddleware) ViewProject(ctx context.Context, token, id string) (project things.Project, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_project for token %s and project %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewProject(ctx, token, id)
}

func (lm *loggingMiddleware) ListProjects(ctx context.Context, token string, offset, limit uint64, name string, metadata things.Metadata) (_ things.ProjectsPage, err error) {
	defer func(begin time.Time) {
		nlog := ""
		if name != "" {
			nlog = fmt.Sprintf("with name %s ", name)
		}
		message := fmt.Sprintf("Method list_projects %sfor token %s took %s to complete", nlog, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListProjects(ctx, token, offset, limit, name, metadata)
}

func (lm *loggingMiddleware) ListProjectsByThing(ctx context.Context, token, id string, offset, limit uint64) (_ things.ProjectsPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_projects_by_thing for thing %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s", message, err))
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListProjectsByThing(ctx, token, id, offset, limit)
}

func (lm *loggingMiddleware) RemoveProject(ctx context.Context, token, id string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove_project for token %s and project %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveProject(ctx, token, id)
}

func (lm *loggingMiddleware) Connect(ctx context.Context, token string, chIDs, thIDs []string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method connect for token %s, projects %s and things %s took %s to complete", token, chIDs, thIDs, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Connect(ctx, token, chIDs, thIDs)
}

func (lm *loggingMiddleware) Disconnect(ctx context.Context, token, projectID, thingID string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method disconnect for token %s, project %s and thing %s took %s to complete", token, projectID, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Disconnect(ctx, token, projectID, thingID)
}

func (lm *loggingMiddleware) CanAccessByKey(ctx context.Context, id, key string) (thing string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method can_access for project %s and thing %s took %s to complete", id, thing, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CanAccessByKey(ctx, id, key)
}

func (lm *loggingMiddleware) CanAccessByID(ctx context.Context, projectID, thingID string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method can_access_by_id for project %s and thing %s took %s to complete", projectID, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CanAccessByID(ctx, projectID, thingID)
}
func (lm *loggingMiddleware) Identify(ctx context.Context, key string) (id string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method identify for key %s and thing %s took %s to complete", key, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Identify(ctx, key)
}
