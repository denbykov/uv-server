package job

import (
	"context"
	"sync"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/messages"
)

type WorkflowAdapter interface {
	CreateWf(
		uuid string,
		config *config.Config,
		ctx context.Context,
		cancel context.CancelFunc,
		wf_in chan interface{},
		wf_out chan interface{},
	)

	RunWf(wg *sync.WaitGroup)
	HandleMessage(message *messages.Message) error
}
