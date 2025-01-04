package handlers

import (
	"errors"
	"fmt"
	"server/business/workflows/download"
	"server/common"
	"server/common/loggers"
	"server/config"
	"server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type DownloadHandler struct {
	log    *logrus.Entry
	config *config.Config
	uuid   *string
}

func NewDownloadHandler(config *config.Config) *DownloadHandler {
	object := &DownloadHandler{}
	object.log = loggers.PresentationLogger
	object.config = config
	object.uuid = nil

	return object
}

func (h *DownloadHandler) Handle(message *messages.Message) error {
	if message.Header.Uuid == nil {
		return errors.New("uuid is required for operation but not specified")
	}

	h.uuid = message.Header.Uuid

	request := &download.Request{}
	err := common.UnmarshalStrict(message.Payload, request)

	if err != nil {
		return fmt.Errorf("failed to parse payload: %v", err)
	}

	controller := download.NewController(h.config)

	err = controller.Run()

	if err != nil {
		return fmt.Errorf("workflow failed: %v", err)
	}

	return nil
}
