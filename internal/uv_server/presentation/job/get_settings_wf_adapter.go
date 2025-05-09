package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"uv_server/internal/uv_protocol"
	jobmessages "uv_server/internal/uv_server/business/data"
	getsettings "uv_server/internal/uv_server/business/workflows/get_settings"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"

	"github.com/sirupsen/logrus"
)

type GetSettingsWfAdapter struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	session_in chan<- *Message
	wf         *getsettings.GetSettingsWf

	resources *data.Resources
}

func NewGetSettingsWfAdapter(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	resources *data.Resources,
) *GetSettingsWfAdapter {
	object := &GetSettingsWfAdapter{}

	object.uuid = uuid
	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "GetSettingsWfAdapter",
			"uuid":      uuid})
	object.config = config
	object.session_in = session_in

	object.resources = resources

	return object
}

func (wa *GetSettingsWfAdapter) CreateWf(
	uuid string,
	config *config.Config,
	ctx context.Context,
	wf_in chan interface{},
	wf_out chan interface{},
) {
	wa.wf = getsettings.NewGetSettingsWf(
		uuid,
		config,
		ctx,
		wf_out,
		wf_in,
		data.NewDatabase(wa.resources.Db),
	)
}

func (wa *GetSettingsWfAdapter) RunWf(
	wg *sync.WaitGroup,
	msg *uv_protocol.Message,
) error {
	if msg.Header.Type != uv_protocol.GetSettingsRequest {
		wa.log.Fatalf("unexpected message type, got %v instead of GetSettingsRequest", msg.Header.Type)
	}

	wg.Add(1)
	go wa.wf.Run(wg)

	return nil
}

func (wa *GetSettingsWfAdapter) HandleSessionMessage(
	msg *uv_protocol.Message,
) error {
	wa.log.Tracef("handling session message: %v", msg.Header.Type)
	message := fmt.Sprintf("unexpected message %v", msg.Header.Type)
	wa.log.Error(message)
	return errors.New(message)
}

func (wa *GetSettingsWfAdapter) HandleWfMessage(
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
					Type: uv_protocol.GetSettingsResponse,
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
