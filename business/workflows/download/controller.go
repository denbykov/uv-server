package download

import (
	"server/common/loggers"
	"server/config"

	"github.com/sirupsen/logrus"
)

type Controller struct {
	log                  *logrus.Entry
	config               *config.Config
	sendProgressMessage  func(*ProgressMessage)
	sendCompletedMessage func(*CompletedMessage)
}

func NewController(
	config *config.Config,
	sendProgressMessage func(*ProgressMessage),
	sendCompletedMessage func(*CompletedMessage),
) *Controller {
	object := &Controller{}
	object.log = loggers.BusinessLogger
	object.config = config
	object.sendProgressMessage = sendProgressMessage
	object.sendCompletedMessage = sendCompletedMessage

	return object
}

func (c *Controller) Run() error {
	progress1 := &ProgressMessage{
		Percentage: 30,
	}

	c.sendProgressMessage(progress1)

	progress2 := &ProgressMessage{
		Percentage: 60,
	}

	c.sendProgressMessage(progress2)

	completed := &CompletedMessage{}

	c.sendCompletedMessage(completed)

	return nil
}
