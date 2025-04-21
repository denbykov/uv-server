package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	getfiles "uv_server/internal/uv_server/business/workflows/get_files"
	jobmessages "uv_server/internal/uv_server/business/workflows/get_files/job_messages"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"

	"github.com/sirupsen/logrus"
)

type GetFilesWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *getfiles.GetFilesWf

	resources *data.Resources
}

func NewGetFilesWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *GetFilesWfAdapter {
	object := &GetFilesWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "GetFilesWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *GetFilesWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = getfiles.NewGetFilesWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		data.NewDatabase(wa.resources.Db),
	)
}

func (wa *GetFilesWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.GetFilesRequest {
		wa.log.Fatalf("unextected message type, got %v instead of GetFilesRequest", msg.Header.Type)
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

func (wa *GetFilesWfAdapter) validateRequest(request *jobmessages.Request) error {
	if request.Limit == nil {
		return fmt.Errorf("missing \"limit\" field")
	}

	if request.Offset == nil {
		return fmt.Errorf("missing \"offset\" field")
	}

	return nil
}

func (wa *GetFilesWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *GetFilesWfAdapter) HandleWfMessage(
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
					Type: uv_protocol.GetFilesResponse,
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
