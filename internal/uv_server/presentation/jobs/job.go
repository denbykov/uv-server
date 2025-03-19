package jobs

import (
	"uv_server/internal/uv_protocol/presentation/messages"
)

type Job interface {
	Run(*messages.Message)
	Notify(*messages.Message)
}

type JobMessage struct {
	Msg  *messages.Message
	Done bool
}
