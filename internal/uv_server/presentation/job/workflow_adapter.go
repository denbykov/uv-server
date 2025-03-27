package job

import (
	"context"
	"sync"
	"uv_server/internal/uv_protocol"
	"uv_server/internal/uv_server/config"
)

type WorkflowAdapter interface {
	CreateWf(
		uuid string,
		config *config.Config,
		jobCtx context.Context,
		wf_in chan interface{},
		wf_out chan interface{},
	)

	RunWf(wg *sync.WaitGroup, msg *uv_protocol.Message) error
	HandleSessionMessage(message *uv_protocol.Message) error
	HandleWfMessage(message interface{}) (State, error)
}
