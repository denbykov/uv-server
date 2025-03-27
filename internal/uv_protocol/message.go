package uv_protocol

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
)

type Type int

const (
	DownloadingRequest Type = iota
	DownloadingProgress
	CancelRequest
	Error
	Done
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

func GetTypes() []string {
	var result []string
	for i := 0; ; i++ {
		t := Type(i)
		str := t.String()
		if str == fmt.Sprintf("Unknown: %d", i) {
			break
		}
		result = append(result, str)
	}
	return result
}

func GetType(name string) (Type, error) {
	var types = GetTypes()
	for index, item := range types {
		if item == name {
			return Type(index), nil
		}
	}
	return 0, errors.New("type is not found")
}

func GetTypeHint() string {
	var validTypes []string
	for _, item := range GetTypes() {
		validTypes = append(validTypes, string(item))
	}
	return strings.Join(validTypes, ", ")
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

func ValidType(type_flag string) (bool, error) {

	allowedTypes := []string{}
	allowedTypes = append(allowedTypes, GetTypes()...)
	ok := slices.Contains(allowedTypes, type_flag)
	if !ok {
		var validTypes []string
		validTypes = append(validTypes, allowedTypes...)
		return false, errors.New("Invalid type. Allowed values: " + strings.Join(validTypes, ", "))
	}
	return true, nil
}
