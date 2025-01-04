package download

import (
	"server/common/loggers"
	"server/config"

	"github.com/sirupsen/logrus"
)

type Controller struct {
	log    *logrus.Entry
	config *config.Config
}

func NewController(config *config.Config) *Controller {
	object := &Controller{}
	object.log = loggers.BusinessLogger
	object.config = config

	return object
}

func (c *Controller) Run() error {

}
