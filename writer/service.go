package writer

import (
	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/errors"
	"github.com/vietquy/alpha/messaging"
	pubsub "github.com/vietquy/alpha/messaging/nats"
	"github.com/vietquy/alpha/transformer"
)

var (
	errOpenConfFile  = errors.New("unable to open configuration file")
	errParseConfFile = errors.New("unable to parse configuration file")
)

// Writer specifies message writing API.
type Writer interface {
	// Write method is used to write received messages.
	// A non-nil error is returned to indicate operation failure.
	Write(messages interface{}) error
}

// Start method starts writing messages received from NATS.
// This method transforms messages before
// using MessageRepository to store them.
func Start(sub messaging.Subscriber, writer Writer, transformer transformer.Transformer, logger logger.Logger) error {
	subjects := []string{pubsub.SubjectAllProjects}

	for _, subject := range subjects {
		if err := sub.Subscribe(subject, handler(transformer, writer)); err != nil {
			return err
		}
	}
	return nil
}

func handler(t transformer.Transformer, c Writer) messaging.MessageHandler {
	return func(msg messaging.Message) error {
		m := interface{}(msg)
		var err error
		if t != nil {
			m, err = t.Transform(msg)
			if err != nil {
				return err
			}
		}
		return c.Write(m)
	}
}