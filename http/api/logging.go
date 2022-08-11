package api

import (
	"context"
	"fmt"
	"time"

	"github.com/vietquy/alpha/http"
	log "github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/messaging"
)

var _ http.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    http.Service
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(svc http.Service, logger log.Logger) http.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Publish(ctx context.Context, token string, msg messaging.Message) (err error) {
	defer func(begin time.Time) {
		destProject := msg.Project
		if msg.Subtopic != "" {
			destProject = fmt.Sprintf("%s.%s", destProject, msg.Subtopic)
		}
		message := fmt.Sprintf("Method publish to project %s took %s to complete", destProject, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Publish(ctx, token, msg)
}
