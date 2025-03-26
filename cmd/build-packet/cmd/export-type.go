package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	msg "uv_server/internal/uv_protocol/presentation/messages"

	"github.com/spf13/cobra"
)

var exportTypeCmd = &cobra.Command{
	Use:   "export-type",
	Short: "Export the available types",
	Long:  `Export the available types to a JavaScript file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Println("Please provide a file path")
			return errors.New("file path is required")
		}
		file_path := args[0]
		if _, err := os.Stat(file_path); os.IsNotExist(err) {
			cmd.Println("Directory does not exist:", file_path)
			return errors.New("directory does not exist")
		}
		if err := GenerateFileJSON(file_path); err != nil {
			cmd.Println("Error generating file:", err)
			return errors.New("error generating file")
		}
		cmd.Println("File generated successfully at", file_path)
		return nil
	},
	Example: `build-packet export-type`,
}

func init() {
	rootCmd.AddCommand(exportTypeCmd)
}

func GenerateFileJSON(file_path string) error {
	types := msg.GetTypes()

	var content strings.Builder
	content.WriteString("const types = {\n")
	for key, item := range types {
		content.WriteString(fmt.Sprintf(`  "%s": %d,`+"\n", item, key))
	}
	content.WriteString("};\n\nexport default types;\n")

	return os.WriteFile(path.Join(file_path, "types.js"), []byte(content.String()), 0644)
}
