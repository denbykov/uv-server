package presentation

import (
	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"uv_server/internal/uv_server/presentation/handlers"
	"uv_server/internal/uv_server/presentation/messages"
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

func (f *HandlerFactory) CreateHandler(
	message *messages.Message) (handlers.Handler, error) {
	typ := message.Header.Type

	f.log.Debugf("Creating handler for message type %v", typ)

	var handler handlers.Handler = nil

	switch typ {
	case messages.Download:
		return handlers.NewDownloadHandler(f.config), nil
	default:
		f.log.Fatalf("Unercognized message type %v", typ)
	}

	return handler, nil
}
