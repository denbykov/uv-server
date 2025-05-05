package job

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	deleteFile "uv_server/internal/uv_server/business/workflows/delete_file"
	jobmessages "uv_server/internal/uv_server/business/workflows/delete_file/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"

	"github.com/sirupsen/logrus"
)

type DeleteFileWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *deleteFile.DeleteFileWf

	resources *data.Resources
}

func NewDeleteFileWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *DeleteFileWfAdapter {
	object := &DeleteFileWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DeleteFileWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *DeleteFileWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = deleteFile.NewDeleteFileWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		data.NewDatabase(wa.resources.Db),
	)
}

func (wa *DeleteFileWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.DeleteFileRequest {
		wa.log.Fatalf("unexpected message type, got %v instead of DeleteFileRequest", msg.Header.Type)
	}

	request := &jobmessages.Request{}
	err := common.UnmarshalStrict(msg.Payload, request)
	if err != nil {
		return fmt.Errorf("failed to parse payload: %v", err)
	}

	err = wa.validateRequest(request)
	if err != nil {
		return fmt.Errorf("request validation failed: %v", err)
	}

	wg.Add(1)
	go wa.wf.Run(wg, request)

	return nil
}

func (wa *DeleteFileWfAdapter) validateRequest(request *jobmessages.Request) error {
	if request.Id == nil {
		return fmt.Errorf("missing \"Id\" field")
	}

	return nil
}

func (wa *DeleteFileWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *DeleteFileWfAdapter) HandleWfMessage(
	msg interface{},
) (State, error) {
	wa.log.Tracef("handling wf message")
	wa.log.Fatalf("Unknown message: %v", reflect.TypeOf(msg))
	return Done, nil
}
