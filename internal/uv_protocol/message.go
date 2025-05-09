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
	DownloadingDone

	CancelRequest
	Error
	Done

	GetFilesRequest
	GetFilesResponse

	GetFileRequest
	GetFileResponse

	UpdateSettingsRequest
	UpdateSettingsResponse
	GetSettingsRequest
	GetSettingsResponse

	Max
)

func (t Type) String() string {
	switch t {
	case DownloadingRequest:
		return "DownloadingRequest"
	case DownloadingProgress:
		return "DownloadingProgress"
	case DownloadingDone:
		return "DownloadingDone"
	case CancelRequest:
		return "CancelRequest"
	case Error:
		return "Error"
	case Done:
		return "Done"
	case GetFilesRequest:
		return "GetFilesRequest"
	case GetFilesResponse:
		return "GetFilesResponse"
	case GetFileRequest:
		return "GetFileRequest"
	case GetFileResponse:
		return "GetFileResponse"
	case UpdateSettingsRequest:
		return "UpdateSettingsRequest"
	case UpdateSettingsResponse:
		return "UpdateSettingsResponse"
	case GetSettingsRequest:
		return "GetSettingsRequest"
	case GetSettingsResponse:
		return "GetSettingsRequest"

	default:
		return fmt.Sprintf("Unknown: %d", t)
	}
}

func GetTypes() []string {
	var types []string
	for t := DownloadingRequest; t < Max; t++ {
		types = append(types, t.String())
	}
	return types
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
	validTypes := GetTypes()
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

	if message.Header.Uuid == nil {
		return nil, fmt.Errorf("message does not contain UUID in the header")
	}

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

func ValidType(name string) (bool, error) {
	types := GetTypes()
	if slices.Contains(types, name) {
		return true, nil
	}
	return false, errors.New("invalid type \"" + name + "\". Allowed values: " + strings.Join(types, ", "))
}
