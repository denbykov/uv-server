package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	msg "uv_server/internal/uv_protocol"

	"github.com/spf13/cobra"
)

var exportTypeCmd = &cobra.Command{
	Use:   "export-type",
	Short: "Export the available types",
	Long:  `Export the available types to a JavaScript file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("file path is required")
		}
		file_path := args[0]
		if _, err := os.Stat(file_path); os.IsNotExist(err) {
			return errors.New("directory does not exist")
		}
		if err := GenerateJSFile(file_path); err != nil {
			return errors.New("error generating file")
		}
		return nil
	},
	Example: `build_packet export-type`,
}

func init() {
	rootCmd.AddCommand(exportTypeCmd)
}

func GenerateJSFile(file_path string) error {
	types := msg.GetTypes()

	var content strings.Builder
	content.WriteString("const types = {\n")
	for key, item := range types {
		content.WriteString(fmt.Sprintf(`  "%s": %d,`+"\n", item, key))
	}
	content.WriteString("};\n\nexport default types;\n")

	return os.WriteFile(path.Join(file_path, "types.js"), []byte(content.String()), 0644)
}
