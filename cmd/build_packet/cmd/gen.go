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

func GenHeader(headerType msg.Type, uuidStr string) (*msg.Header, error) {

	if uuidStr == "" {
		newUuid := uuid.New().String()
		uuidStr = newUuid
	}
	fmt.Printf("UUID: %s\n", uuidStr)
	return &msg.Header{
		Type: headerType,
		Uuid: &uuidStr,
	}, nil
}

func GenMessage(headerType msg.Type, uuidStr string, payload []byte) (*msg.Message, error) {
	header, err := GenHeader(headerType, uuidStr)
	if err != nil {
		return nil, err
	}

	return &msg.Message{
		Header:  header,
		Payload: payload,
	}, nil
}

func GenHexdump(headerType msg.Type, uuidStr string, data []byte) (string, error) {
	msg, err := GenMessage(headerType, uuidStr, data)
	if err != nil {
		return "", err
	}

	serialized := msg.Serialize()
	hexdump := hex.EncodeToString(serialized)

	return hexdump, nil
}
