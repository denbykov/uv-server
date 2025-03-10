package cmd

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Path   string
	Header string
	Data   string
)

var inputCmd = &cobra.Command{
	Use:   "input",
	Short: "Generate a hexdump package",
	Long: `This command generates a hexdump representation of given header and data.
If a path is provided, it saves the output to a file. Otherwise, it prints to terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Header == "" || Data == "" {
			return errors.New("both --header and --data flags are required")
		}

		var headerJSON, dataJSON json.RawMessage
		result := make([]byte, 0)

		if err := json.Unmarshal([]byte(Header), &headerJSON); err != nil {
			return fmt.Errorf("failed to parse header: %v", err)
		}
		if err := json.Unmarshal([]byte(Data), &dataJSON); err != nil {
			return fmt.Errorf("failed to parse data: %v", err)
		}

		headerBytes, _ := json.Marshal(headerJSON)
		dataBytes, _ := json.Marshal(dataJSON)

		result = binary.BigEndian.AppendUint32(result, uint32(len(headerBytes)))
		result = slices.Concat(result, headerBytes, dataBytes)
		hexdump := fmt.Sprintf("%x", result)

		if Path != "" {
			return os.WriteFile(Path, []byte(hexdump), 0644)
		}

		fmt.Println(hexdump)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inputCmd)

	inputCmd.PersistentFlags().StringVarP(&Path, "path", "p", "", "Define the path to save hexdump package.")
	inputCmd.PersistentFlags().StringVar(&Header, "header", "", "JSON formatted header (required)")
	inputCmd.PersistentFlags().StringVar(&Data, "data", "", "JSON formatted data (required)")

	inputCmd.MarkPersistentFlagRequired("header")
	inputCmd.MarkPersistentFlagRequired("data")

	viper.BindPFlag("path", inputCmd.PersistentFlags().Lookup("path"))
}
