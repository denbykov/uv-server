package download

import (
	"server/business/workflows/download/client"
	"server/business/workflows/download/data"
	"server/common/loggers"
	"server/config"

	"github.com/sirupsen/logrus"
)

type Controller struct {
	log                 *logrus.Entry
	config              *config.Config
	sendProgressMessage func(*client.ProgressMessage)
	sendDoneMessage     func(*client.DoneMessage)

	downloader data.Downloader
}

func NewController(
	config *config.Config,
	sendProgressMessage func(*client.ProgressMessage),
	sendDoneMessage func(*client.DoneMessage),
	downloader data.Downloader,
) *Controller {
	object := &Controller{}
	object.log = loggers.BusinessLogger.WithField(
		"component", "DownloadingController")
	object.config = config
	object.sendProgressMessage = sendProgressMessage
	object.sendDoneMessage = sendDoneMessage
	object.downloader = downloader

	object.downloader.RegisterOnProgress(object.onProgress)

	return object
}

func (c *Controller) onProgress(msg *data.ProgressMessage) {
	progress1 := &client.ProgressMessage{
		Percentage: msg.Percentage,
	}

	c.sendProgressMessage(progress1)
}

func (c *Controller) Run(url string) {
	c.log.Debugf("Downloading file: %v", url)

	filename, err := c.downloader.Download(url)

	if err != nil {
		c.log.Errorf("Donwloading failed: %v", err)
	}

	_ = filename

	c.sendDoneMessage(&client.DoneMessage{})

	c.log.Debugf("Downloading done for file: %v", url)
}
