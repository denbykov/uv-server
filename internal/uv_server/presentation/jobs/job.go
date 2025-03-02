package jobs

import (
	"context"
	"uv_server/internal/uv_server/presentation/messages"
)

type Job interface {
	Run(context.Context, *messages.Message) error
	Notify(*messages.Message)
}
