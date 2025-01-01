package presentation

import (
	"server/common/loggers"
	"server/config"
	"server/presentation/handlers"
	"server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type HandlerFactory struct {
	log    *logrus.Entry
	config *config.Config
}

func NewHandlerFactory(config *config.Config) *HandlerFactory {
	object := &HandlerFactory{}
	object.log = loggers.PresentationLogger
	object.config = config

	return object
}

func (f *HandlerFactory) CreateHandler(message *messages.Message) handlers.Handler {
	typ := message.Header.Type

	f.log.Debugf("Creating handler for message type %v", typ)

	var handler handlers.Handler = nil

	switch typ {
	case messages.Download:
		handler = nil
	default:
		f.log.Fatalf("Unercognized message type %v", typ)
	}

	return handler
}
