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
	Download          Type = 1
	DownloadCompleted Type = 10
	DownloadProgress  Type = 11
)

func (t Type) String() string {
	switch t {
	case Download:
		return "Download"
	case DownloadCompleted:
		return "DownloadCompleted"
	case DownloadProgress:
		return "DownloadProgress"
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
