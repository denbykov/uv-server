package cjmessages

type Error struct {
	Reason string `json:"reason"`
}

var InternalError = Error{Reason: "Internal error"}

type Done struct {
}
