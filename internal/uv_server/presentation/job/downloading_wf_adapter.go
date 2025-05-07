package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	"uv_server/internal/uv_server/business/workflows/downloading"
	jobmessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"
	"uv_server/internal/uv_server/data/downloaders"

	"github.com/sirupsen/logrus"
)

type DownloadingWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *downloading.DownloadingWf

	downloaderOut chan interface{}

	resources *data.Resources
}

func NewDownloadingWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *DownloadingWfAdapter {
	object := &DownloadingWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *DownloadingWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.downloaderOut = make(chan interface{}, 1)

	downloader := downloaders.NewYtDownloader(
		uuid,
		config,
		ctx,
		wa.downloaderOut,
		wa.resources.To_clean,
	)

	wa.wf = downloading.NewDownloadingWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		downloader,
		wa.downloaderOut,
		data.NewDatabase(wa.resources.Db),
	)
}

func (wa *DownloadingWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.DownloadingRequest {
		wa.log.Fatalf("unexpected message type, got %v instead of DownloadingRequest", msg.Header.Type)
	}

	request := &jobmessages.Request{}
	err := common.UnmarshalStrict(msg.Payload, request)

	if err != nil {
		message := fmt.Sprintf("failed to parse payload: %v", err)
		wa.log.Errorf(message)
		return errors.New(message)
	}

	wg.Add(1)
	go wa.wf.Run(wg, request)

	return nil
}

func (wa *DownloadingWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *DownloadingWfAdapter) HandleWfMessage(
	msg interface{},
) (State, error) {
	wa.log.Tracef("handling wf message")

	if tMsg, ok := msg.(*jobmessages.Progress); ok {
		payload, err := json.Marshal(tMsg)
		if err != nil {
			wa.log.Fatalf("failed to serialize message: %v", err)
		}

		msg := &Message{
			Msg: &uv_protocol.Message{
				Header: &uv_protocol.Header{
					Uuid: &wa.uuid,
					Type: uv_protocol.DownloadingProgress,
				},
				Payload: payload,
			},
			Done: false,
		}

		wa.session_in <- msg
	} else if tMsg, ok := msg.(*jobmessages.Done); ok {
		payload, err := json.Marshal(tMsg)
		if err != nil {
			wa.log.Fatalf("failed to serialize message: %v", err)
		}

		msg := &Message{
			Msg: &uv_protocol.Message{
				Header: &uv_protocol.Header{
					Uuid: &wa.uuid,
					Type: uv_protocol.DownloadingDone,
				},
				Payload: payload,
			},
			Done: true,
		}

		wa.session_in <- msg

		return Done, nil
	} else {
		wa.log.Fatalf("Unknown message: %v", reflect.TypeOf(msg))
	}

	return Active, nil
}
