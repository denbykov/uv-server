package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	msg "uv_server/internal/uv_protocol/presentation/messages"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	Payload string
	Type    string
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a hexdump packet",
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
		packetType, _ := strconv.Atoi(Type)
		packet, err := GenHexdump(packetType, []byte(Payload))
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
}

func GenHeader(header_type int) (*msg.Header, error) {
	uuidStr := uuid.New().String()
	return &msg.Header{
		Type: msg.Type(header_type),
		Uuid: &uuidStr,
	}, nil
}

func GenMessage(header_type int, payload []byte) (*msg.Message, error) {
	header, err := GenHeader(header_type)
	if err != nil {
		return nil, err
	}

	return &msg.Message{
		Header:  header,
		Payload: payload,
	}, nil
}

func GenHexdump(header_type int, data []byte) (string, error) {
	msg, err := GenMessage(header_type, data)
	if err != nil {
		return "", err
	}

	serialized := msg.Serialize()
	hexdump := hex.EncodeToString(serialized)

	return hexdump, nil
}
