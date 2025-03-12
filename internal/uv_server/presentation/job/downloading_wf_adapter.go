package job

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_server/business/workflows/downloading"
	jobmessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"
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

	downloaderOut chan interface{}

	db *sql.DB
}

func NewDownloadingWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	db *sql.DB,
) *DownloadingWfAdapter {
	object := &DownloadingWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.db = db

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
	)

	wa.wf = downloading.NewDownloadingWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		downloader,
		wa.downloaderOut,
		data.NewDatabase(wa.db),
	)
}

func (wa *DownloadingWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *messages.Message,
) error {
	request := &jobmessages.Request{}
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

	if tMsg, ok := msg.(*jobmessages.Progress); ok {
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
