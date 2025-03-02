package handlers

import "uv_server/internal/uv_server/presentation/messages"

type Handler interface {
	Handle(*messages.Message, chan *messages.Message) error
}
