package api

import (
	"fmt"
	"time"

	"github.com/vietquy/alpha/writer"
	log "github.com/vietquy/alpha/logger"
)

var _ writer.Writer = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger   log.Logger
	writer 	 writer.Writer
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(writer writer.Writer, logger log.Logger) writer.Writer {
	return &loggingMiddleware{
		logger:   logger,
		writer:   writer,
	}
}

func (lm *loggingMiddleware) Write(msgs interface{}) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method write took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.writer.Write(msgs)
}
