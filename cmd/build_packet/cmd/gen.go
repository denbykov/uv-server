package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"

	msg "uv_server/internal/uv_protocol"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	Payload string
	Type    string
	Uuid    string
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a packet hexdump",
	Long:  `Encode JSON header and data into a binary hexdump.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Type == "" {
			return errors.New("type is required")
		}
		if Payload == "" {
			return errors.New("payload is required")
		}
		valid, err := msg.ValidType(Type)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("invalid type")
		}
		packetType, err := msg.GetType(Type)
		if err != nil {
			return err
		}
		packet, err := GenHexdump(packetType, Uuid, []byte(Payload))
		if err != nil {
			return err
		}
		fmt.Println(packet)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.PersistentFlags().StringVarP(&Payload, "payload", "p", "", "JSON formatted payload (required)")
	genCmd.PersistentFlags().StringVarP(&Type, "type", "t", "", "Type of the packet (required). Available: "+msg.GetTypeHint())
	genCmd.PersistentFlags().StringVarP(&Uuid, "uuid", "u", "", "Enter your own Universally Unique Identifier (UUID) (not required)")
}

func GenHeader(header_type msg.Type, _uuid string) (*msg.Header, error) {

	if _uuid == "" {
		newUuid := uuid.New().String()
		_uuid = newUuid
	}
	fmt.Printf("UUID: %s\n", _uuid)
	return &msg.Header{
		Type: header_type,
		Uuid: &_uuid,
	}, nil
}

func GenMessage(header_type msg.Type, _uuid string, payload []byte) (*msg.Message, error) {
	header, err := GenHeader(header_type, _uuid)
	if err != nil {
		return nil, err
	}

	return &msg.Message{
		Header:  header,
		Payload: payload,
	}, nil
}

func GenHexdump(header_type msg.Type, _uuid string, data []byte) (string, error) {
	msg, err := GenMessage(header_type, _uuid, data)
	if err != nil {
		return "", err
	}

	serialized := msg.Serialize()
	hexdump := hex.EncodeToString(serialized)

	return hexdump, nil
}
