package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "build-packet",
	Short: "This utility will generate a hexdump packet for testing uv_server ",
	Long: `The "build-packet" utility is a powerful tool designed to generate hexdump packages 
    for testing uv_server. This tool allows developers and system administrators to create structured 
    binary test packages for debugging, performance benchmarking, and protocol verification.
    
    Usage:
    - Generate a hexdump packet from a string.
    
    Examples:
    1. Generate a hexdump packet from a date:
       build-packet -gen -t DownloadingRequest -p {"url": "http://example.com"}
        
    By leveraging this tool, developers can efficiently create test cases for uv_server, ensuring 
    robustness, reliability, and correctness in handling binary data streams.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
