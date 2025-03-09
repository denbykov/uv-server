package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_server/business/workflows/downloading"
	jobMessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data/downloaders"
	"uv_server/internal/uv_server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type DownloadingWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *downloading.DownloadingWf

	downloader_out chan interface{}
}

func NewDownloadingWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
) *DownloadingWfAdapter {
	object := &DownloadingWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	return object
}

func (wa *DownloadingWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.downloader_out = make(chan interface{}, 1)

	downloader := downloaders.NewYtDownloader(
		config,
		ctx,
		wa.downloader_out,
	)

	wa.wf = downloading.NewDownloadingWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		downloader,
		wa.downloader_out,
	)
}

func (wa *DownloadingWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *messages.Message,
) error {
	request := &jobMessages.Request{}
	err := common.UnmarshalStrict(msg.Payload, request)

	if err != nil {
		return fmt.Errorf("failed to parse payload: %v", err)
	}

	wg.Add(1)
	go wa.wf.Run(wg, request)

	return nil
}

func (wa *DownloadingWfAdapter) HandleSessionMessage(
	msg *messages.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *DownloadingWfAdapter) HandleWfMessage(
	msg interface{},
) (State, error) {
	wa.log.Tracef("handling wf message")

	if tMsg, ok := msg.(jobMessages.Progress); ok {
		payload, err := json.Marshal(tMsg)
		if err != nil {
			wa.log.Fatalf("failed to serialize message: %v", err)
		}

		msg := &Message{
			Msg: &messages.Message{
				Header: &messages.Header{
					Uuid: &wa.uuid,
					Type: messages.DownloadingProgress,
				},
				Payload: payload,
			},
			Done: false,
		}

		wa.session_in <- msg
	} else {
		wa.log.Fatalf("Unknown message: %v", reflect.TypeOf(msg))
	}

	return Active, nil
}
