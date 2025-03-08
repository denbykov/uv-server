package messages

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
)

type Type int

const (
	DownloadingRequest  Type = 1
	DownloadingProgress Type = 10
	CancelRequest       Type = 2
	Error               Type = 3
	Done                Type = 4
)

func (t Type) String() string {
	switch t {
	case DownloadingRequest:
		return "DownloadingRequest"
	case DownloadingProgress:
		return "DownloadingProgress"
	case CancelRequest:
		return "CancelRequest"
	case Error:
		return "Error"
	case Done:
		return "Done"
	default:
		return fmt.Sprintf("Unknown: %d", t)
	}
}

type Header struct {
	Type Type    `json:"type"`
	Uuid *string `json:"uuid"`
}

type Message struct {
	Header  *Header
	Payload []byte
}

func ParseMessage(data []byte) (*Message, error) {
	log := loggers.PresentationLogger

	log.Debugf("Message size: %v", len(data))

	message := &Message{
		Header:  &Header{},
		Payload: nil,
	}

	var offset int = 0

	if offset+4 > len(data) {
		return nil, errors.New("message does not include header size")
	}

	headerSize := binary.BigEndian.Uint32(data[offset:4])
	offset += 4

	if offset+int(headerSize) > len(data) {
		return nil, errors.New("message does not include header")
	}

	header := data[offset : offset+int(headerSize)]
	offset += int(headerSize)

	log.Tracef("Header: %v", string(header))

	err := common.UnmarshalStrict(header, message.Header)

	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %v", err)
	}

	message.Payload = data[offset:]
	offset += int(headerSize)

	_ = offset

	return message, nil
}

func (m *Message) Serialize() []byte {
	log := loggers.PresentationLogger

	result := make([]byte, 0)

	header, err := json.Marshal(m.Header)
	if err != nil {
		log.Fatalf("Failed to serialize message: %v", err)
	}

	result = binary.BigEndian.AppendUint32(result, uint32(len(header)))
	result = slices.Concat(result, header, m.Payload)

	return result
}
