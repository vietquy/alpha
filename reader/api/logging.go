package api

import (
	"fmt"
	"time"

	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/reader"
)

var _ reader.MessageRepository = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger logger.Logger
	svc    reader.MessageRepository
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc reader.MessageRepository, logger logger.Logger) reader.MessageRepository {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}

func (lm *loggingMiddleware) ReadAll(projectID string, rpm reader.PageMetadata) (page reader.MessagesPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method read_all for project %s with query %v took %s to complete", projectID, rpm, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ReadAll(projectID, rpm)
}
