package job

import (
	"context"
	"fmt"
	"sync"
	"uv_server/internal/uv_server/business/workflows/downloading"
	jobmessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type DownloadingWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	wf *downloading.DownloadingWf
}

func NewDownloadingWfAdapter(
	uuid string,
	config *config.Config,
) *DownloadingWfAdapter {
	object := &DownloadingWfAdapter{}

	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.uuid = uuid

	return object
}

func (wa *DownloadingWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	cancel context.CancelFunc,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = downloading.NewDownloadingWf(
		uuid,
		config,
		ctx,
		cancel,
	)
}

func (wa *DownloadingWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *messages.Message,
) error {
	defer wg.Done()

	request := &jobmessages.Request{}
	err := common.UnmarshalStrict(msg.Payload, request)

	if err != nil {
		return fmt.Errorf("failed to parse payload: %v", err)
	}

	wg.Add(1)
	go wa.wf.Run(request)

	return nil
}

func (wa *DownloadingWfAdapter) HandleMessage(
	message *messages.Message,
) error {
	wa.log.Tracef("handling message: %v", message.Header.Type)
	return fmt.Errorf("unexpected message %v", message.Header.Type)
}
