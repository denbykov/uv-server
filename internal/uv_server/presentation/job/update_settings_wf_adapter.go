package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	jobmessages "uv_server/internal/uv_server/business/data"
	updatesettings "uv_server/internal/uv_server/business/workflows/update_settings"
	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"

	"github.com/sirupsen/logrus"
)

type UpdateSettingsWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *updatesettings.UpdateSettingsWf

	resources *data.Resources
}

func NewSetSettingsWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *UpdateSettingsWfAdapter {
	object := &UpdateSettingsWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "UpdateSettingsWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *UpdateSettingsWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = updatesettings.NewUpdateSettingsWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		data.NewDatabase(wa.resources.Db),
	)
}

func (wa *UpdateSettingsWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.UpdateSettingsRequest {
		wa.log.Fatalf("unexpected message type, got %v instead of UpdateSettingsRequest", msg.Header.Type)
	}

	request := &jobmessages.Settings{}
	err := common.UnmarshalStrict(msg.Payload, request)
	if err != nil {
		return fmt.Errorf("failed to parse payload: %v", err)
	}

	wg.Add(1)
	go wa.wf.Run(wg, request)

	return nil
}

func (wa *UpdateSettingsWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	return fmt.Errorf("unexpected message %v", msg.Header.Type)
}

func (wa *UpdateSettingsWfAdapter) HandleWfMessage(
	msg interface{},
) (State, error) {
	wa.log.Tracef("handling wf message")

	if tMsg, ok := msg.(*jobmessages.Settings); ok {
		payload, err := json.Marshal(tMsg)
		if err != nil {
			wa.log.Fatalf("failed to serialize message: %v", err)
		}

		msg := &Message{
			Msg: &uv_protocol.Message{
				Header: &uv_protocol.Header{
					Uuid: &wa.uuid,
					Type: uv_protocol.UpdateSettingsResponse,
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
