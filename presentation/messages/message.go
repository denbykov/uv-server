package messages

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"server/common/loggers"
	"slices"
)

type Type int

const (
	Download Type = 0
)

type Header struct {
	Type Type
}

type Message struct {
	Header  *Header
	Payload []byte
}

func ParseMessage(data []byte) *Message {
	log := loggers.PresentationLogger

	message := &Message{}

	var offset int = 0

	headerSize := binary.LittleEndian.Uint32(data[offset:4])
	offset += 4

	header := data[offset : offset+int(headerSize)]
	offset += int(headerSize)

	err := json.Unmarshal(header, &message.Header)

	if err != nil {
		log.Fatalf("Failed to parse message: %v", err)
	}

	message.Payload = data[offset:]
	offset += int(headerSize)

	return message
}

func (m *Message) Serialize() []byte {
	result := make([]byte, 0)

	header, err := json.Marshal(m.Header)
	if err != nil {
		log.Fatalf("Failed to serialize message: %v", err)
	}

	result = binary.BigEndian.AppendUint32(result, uint32(len(header)))
	result = slices.Concat(result, header, m.Payload)

	return result
}
