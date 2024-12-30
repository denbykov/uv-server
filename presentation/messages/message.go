package messages

type Type int

const (
	Download Type = 0
)

type Header struct {
	Type Type
}

type Message interface {
	Size() int32
	Header() *Header
}
