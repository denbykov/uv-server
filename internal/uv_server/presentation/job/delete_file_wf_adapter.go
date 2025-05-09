package job

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	deleteFiles "uv_server/internal/uv_server/business/workflows/delete_files"
	jobmessages "uv_server/internal/uv_server/business/workflows/delete_files/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"

	"github.com/sirupsen/logrus"
)

type DeleteFilesWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *deleteFiles.DeleteFilesWf

	resources *data.Resources
}

func NewDeleteFilesWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *DeleteFilesWfAdapter {
	object := &DeleteFilesWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DeleteFilesWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *DeleteFilesWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = deleteFiles.NewDeleteFilesWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		data.NewDatabase(wa.resources.Db),
		data.NewFilesystem(),
	)
}

func (wa *DeleteFilesWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.DeleteFilesRequest {
		wa.log.Fatalf("unexpected message type, got %v instead of DeleteFilesRequest", msg.Header.Type)
	}

	request := &jobmessages.Request{}
	err := common.UnmarshalStrict(msg.Payload, request)
	if err != nil {
		newErr := fmt.Errorf("failed to parse payload: %w", err)
		wa.log.Error(newErr)
		return newErr
	}

	err = wa.validateRequest(request)
	if err != nil {
		newErr := fmt.Errorf("request validation failed: %v", err)
		wa.log.Error(newErr)
		return newErr
	}

	wg.Add(1)
	go wa.wf.Run(wg, request)

	return nil
}

func (wa *DeleteFilesWfAdapter) validateRequest(request *jobmessages.Request) error {
	if len(request.Ids) == 0 {
		return fmt.Errorf("\"Ids\" array is empty")
	}

	return nil
}

func (wa *DeleteFilesWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *DeleteFilesWfAdapter) HandleWfMessage(
	msg interface{},
) (State, error) {
	wa.log.Tracef("handling wf message")
	wa.log.Fatalf("Unknown message: %v", reflect.TypeOf(msg))
	return Done, nil
}
