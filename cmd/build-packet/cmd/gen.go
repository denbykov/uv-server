package cmd

import (
	"errors"
	"fmt"
	"strconv"

	gen "uv_server/internal/build-packet/presentation"
	msg "uv_server/internal/uv_protocol/presentation/messages"

	"github.com/spf13/cobra"
)

var (
	Path     string
	Data     string
	TypeFlag string
	Hexdump  bool
	JSTFile  bool
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a hexdump packet",
	Long: `Encode JSON header and data into a binary hexdump.
If a file path is specified, the output is saved; otherwise, it is printed to the terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("jst") {
			if Path == "" {
				return errors.New("path is required")
			}
			return msg.GenerateJSFile(Path)
		}
		if cmd.Flags().Changed("hexdump") {
			if TypeFlag == "" {
				return errors.New("type is required")
			}
			if Data == "" {
				return errors.New("data is required")
			}
			valid, err := msg.ValidType(TypeFlag)
			if err != nil {
				return err
			}
			if !valid {
				return errors.New("invalid type")
			}

			packetType, _ := strconv.Atoi(TypeFlag)
			packet, err := gen.GenHexdumpPacket(packetType, []byte(Data))

			if Path != "" {
				return gen.SaveHexdumpPacket(packet, Path)
			}

			if err != nil {
				return err
			}
			fmt.Println(packet)
			return nil
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.PersistentFlags().StringVarP(&Path, "path", "p", "", "Define the path to save hexdump packet.")
	genCmd.PersistentFlags().StringVarP(&Data, "data", "d", "", "JSON formatted data (required)")
	genCmd.PersistentFlags().StringVarP(&TypeFlag, "type", "t", "", "Type of the packet (required). Available: "+msg.GetTypeHint())
	genCmd.PersistentFlags().BoolVarP(&Hexdump, "hexdump", "x", false, "Print hexdump to the terminal")
	genCmd.PersistentFlags().BoolVarP(&JSTFile, "jst", "j", false, "Path to the JS Type file")
}
