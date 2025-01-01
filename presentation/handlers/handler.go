package handlers

import "server/presentation/messages"

type Handler interface {
	Handle(*messages.Message)
}
