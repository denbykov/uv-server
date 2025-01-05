package download

import (
	"server/business/workflows/download/client"
	"server/common/loggers"
	"server/config"

	"github.com/sirupsen/logrus"
)

type Controller struct {
	log                  *logrus.Entry
	config               *config.Config
	sendProgressMessage  func(*client.ProgressMessage)
	sendCompletedMessage func(*client.CompletedMessage)
}

func NewController(
	config *config.Config,
	sendProgressMessage func(*client.ProgressMessage),
	sendCompletedMessage func(*client.CompletedMessage),
) *Controller {
	object := &Controller{}
	object.log = loggers.BusinessLogger
	object.config = config
	object.sendProgressMessage = sendProgressMessage
	object.sendCompletedMessage = sendCompletedMessage

	return object
}

func (c *Controller) Run() error {
	// progress1 := &client.ProgressMessage{
	// 	Percentage: 30,
	// }

	// c.sendProgressMessage(progress1)

	// progress2 := &client.ProgressMessage{
	// 	Percentage: 60,
	// }

	// c.sendProgressMessage(progress2)

	// completed := &client.CompletedMessage{}

	// c.sendCompletedMessage(completed)

	return nil
}
