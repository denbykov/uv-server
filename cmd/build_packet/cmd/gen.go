package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"

	msg "uv_server/internal/uv_protocol"

	"slices"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	Payload *string
	Type    string
	Uuid    string
)

var emptyPayloadTypes = []msg.Type{
	msg.GetSettingsRequest,
	msg.GetSettingsResponse,
}

func isEmptyPayloadType(msgType msg.Type) bool {
	return slices.Contains(emptyPayloadTypes, msgType)
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a packet hexdump",
	Long:  `Encode JSON header and data into a binary hexdump.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Type == "" {
			return errors.New("type is required")
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

		var payloadData []byte
		if Payload == nil {
			if isEmptyPayloadType(packetType) {
				payloadData = nil
			} else {
				return errors.New("payload is required")
			}
		} else {
			payloadData = []byte(*Payload)
		}

		packet, err := GenHexdump(packetType, Uuid, payloadData)
		if err != nil {
			return err
		}
		fmt.Println(packet)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	var emptyPayload string
	genCmd.PersistentFlags().StringVarP(&emptyPayload, "payload", "p", "", "JSON formatted payload (required for non-empty payload types)")
	if cmd, _, err := rootCmd.Find([]string{"gen"}); err == nil {
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			if flag := cmd.Flag("payload"); flag != nil && flag.Changed {
				value := flag.Value.String()
				Payload = &value
			}
			return nil
		}
	}
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
