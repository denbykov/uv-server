package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	getfile "uv_server/internal/uv_server/business/workflows/get_file"
	jobmessages "uv_server/internal/uv_server/business/workflows/get_file/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"

	"github.com/sirupsen/logrus"
)

type GetFileWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *getfile.GetFileWf

	resources *data.Resources
}

func NewGetFileWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *GetFileWfAdapter {
	object := &GetFileWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "GetFileWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *GetFileWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = getfile.NewGetFileWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		data.NewDatabase(wa.resources.Db),
	)
}

func (wa *GetFileWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.GetFileRequest {
		wa.log.Fatalf("unexpected message type, got %v instead of GetFileRequest", msg.Header.Type)
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

func (wa *GetFileWfAdapter) validateRequest(request *jobmessages.Request) error {
	if request.Id == nil {
		return fmt.Errorf("missing \"Id\" field")
	}

	return nil
}

func (wa *GetFileWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *GetFileWfAdapter) HandleWfMessage(
	msg interface{},
) (State, error) {
	wa.log.Tracef("handling wf message")

	if tMsg, ok := msg.(*jobmessages.Result); ok {
		payload, err := json.Marshal(tMsg)
		if err != nil {
			wa.log.Fatalf("failed to serialize message: %v", err)
		}

		msg := &Message{
			Msg: &uv_protocol.Message{
				Header: &uv_protocol.Header{
					Uuid: &wa.uuid,
					Type: uv_protocol.GetFileResponse,
				},
				Payload: payload,
			},
			Done: true,
		}

		wa.session_in <- msg
	} else {
		wa.log.Fatalf("Unknown message: %v", reflect.TypeOf(msg))
	}

	return Done, nil
}
