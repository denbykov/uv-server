package presentation

import (
	"server/common/loggers"
	"server/config"

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

func (h *HandlerFactory) 