package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"server/business/workflows/download"
	download_messages "server/business/workflows/download/client"
	"server/common"
	"server/common/loggers"
	"server/config"
	"server/data"
	"server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type DownloadHandler struct {
	log    *logrus.Entry
	config *config.Config
	uuid   *string
	send   chan *messages.Message
}

func NewDownloadHandler(config *config.Config) *DownloadHandler {
	object := &DownloadHandler{}
	object.log = loggers.PresentationLogger.WithField(
		"component", "DownloadHandler")
	object.config = config
	object.uuid = nil

	return object
}

func (h *DownloadHandler) sendDoneMessage(
	message *download_messages.DoneMessage) {
	h.log.Tracef("Sending done message %v", message)

	payload, err := json.Marshal(message)
	if err != nil {
		h.log.Fatalf("Failed to serialize message: %v", err)
	}

	msg := &messages.Message{
		Header: &messages.Header{
			Type: messages.DownloadCompleted,
			Uuid: h.uuid,
		},
		Payload: payload,
	}

	h.send <- msg
}

func (h *DownloadHandler) sendProgressMessage(
	message *download_messages.ProgressMessage) {
	h.log.Tracef("Sending progress message %v", message)

	payload, err := json.Marshal(message)
	if err != nil {
		h.log.Fatalf("Failed to serialize message: %v", err)
	}

	msg := &messages.Message{
		Header: &messages.Header{
			Type: messages.DownloadProgress,
			Uuid: h.uuid,
		},
		Payload: payload,
	}

	h.send <- msg
}

func (h *DownloadHandler) Handle(
	message *messages.Message,
	send chan *messages.Message,
) error {
	h.log.Debugf("Handling message %v", message)

	h.send = send

	if message.Header.Uuid == nil {
		return errors.New("uuid is required for operation but not specified")
	}

	h.uuid = message.Header.Uuid

	request := &download_messages.Request{}
	err := common.UnmarshalStrict(message.Payload, request)

	if err != nil {
		return fmt.Errorf("failed to parse payload: %v", err)
	}

	controller := download.NewController(
		h.config,
		h.sendProgressMessage,
		h.sendDoneMessage,
		data.NewDownloader(h.config),
	)

	controller.Run(*request.Url)

	h.log.Debugf("Message handling completed %v", message)

	return nil
}
