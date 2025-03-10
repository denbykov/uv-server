/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "package-creation",
	Short: "This utility will generate a hexdump package for testing uv_server ",
	Long: `The "package-creation" utility is a powerful tool designed to generate hexdump packages 
    for testing uv_server. This tool allows developers and system administrators to create structured 
    binary test packages for debugging, performance benchmarking, and protocol verification.
    
    Usage:
    - Generate a hexdump package from an input file or string.
    - Customize the output format and structure to fit specific testing needs.
    - Integrate with automated test suites to validate server behavior under various conditions.
    
    Examples:
    1. Generate a hexdump package from a file:
       package-creation -input file.bin -output test_package.hex
    
    2. Create a test package from a string:
       package-creation -data "Hello, world!" -output test_package.hex
    
    3. Customize the output with additional options:
       package-creation -input file.bin -format compact -compress -output optimized_package.hex
    
    By leveraging this tool, developers can efficiently create test cases for uv_server, ensuring 
    robustness, reliability, and correctness in handling binary data streams.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.package-creation.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
