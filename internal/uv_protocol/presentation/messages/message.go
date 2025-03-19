package messages

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
)

type Type int

const (
	DownloadingRequest Type = iota + 1
	DownloadingCompleted
	DownloadingProgress
	DownloadingCancelled
)

func (t Type) String() string {
	switch t {
	case DownloadingRequest:
		return "DownloadingRequest"
	case DownloadingCompleted:
		return "DownloadingCompleted"
	case DownloadingProgress:
		return "DownloadingProgress"
	case DownloadingCancelled:
		return "DownloadingCancelled"
	default:
		return fmt.Sprintf("Unknown: %d", t)
	}
}

func GetTypes() []map[string]string {
	var result []map[string]string

	for i := 1; ; i++ {
		t := Type(i)
		str := t.String()
		if str == fmt.Sprintf("Unknown: %d", i) {
			break
		}
		result = append(result, map[string]string{
			fmt.Sprintf("%d", i): str,
		})
	}

	return result
}

func GetTypeHint() string {
	var validTypes []string
	for _, item := range GetTypes() {
		for key, value := range item {
			validTypes = append(validTypes, fmt.Sprintf("%s (%s)", key, value))
		}
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

func GenerateJSFile(filename string) error {
	types := GetTypes()

	var content strings.Builder
	content.WriteString("const types = {\n")
	for _, item := range types {
		for key, value := range item {
			content.WriteString(fmt.Sprintf(`  "%s": "%s",`+"\n", key, value))
		}
	}
	content.WriteString("};\n\nexport default types;\n")

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func ValidType(type_flag string) (bool, error) {

	allowedTypes := make(map[string]string)
	for _, item := range GetTypes() {
		maps.Copy(allowedTypes, item)
	}
	_, ok := allowedTypes[type_flag]
	if !ok {
		var validTypes []string
		for key, value := range allowedTypes {
			validTypes = append(validTypes, fmt.Sprintf("%s (%s)", key, value))
		}
		return false, errors.New("Invalid type. Allowed values: " + strings.Join(validTypes, ", "))
	}
	return true, nil
}
