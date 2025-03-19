package presentation

import (
	"encoding/hex"
	"os"
	msg "uv_server/internal/uv_protocol/presentation/messages"
)

func GenHeaderPacket(header_type int) (*msg.Header, error) {
	return &msg.Header{
		Type: msg.Type(header_type),
		Uuid: nil,
	}, nil
}

func GenMessagePacket(header_type int, data []byte) (*msg.Message, error) {
	header, err := GenHeaderPacket(header_type)
	if err != nil {
		return nil, err
	}

	return &msg.Message{
		Header:  header,
		Payload: data,
	}, nil
}

func GenHexdumpPacket(header_type int, data []byte) (string, error) {
	msg, err := GenMessagePacket(header_type, data)
	if err != nil {
		return "", err
	}

	serialized := msg.Serialize()
	hexdump := hex.EncodeToString(serialized)

	return hexdump, nil
}

func SaveHexdumpPacket(hexdump string, filename string) error {
	os.WriteFile(filename, []byte(hexdump), 0644)
	return nil
}
