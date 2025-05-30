package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// NewRootCmd creates and returns the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication service",
		Long:  "Authentication service",
	}

	// Add commands
	rootCmd.AddCommand(NewServeCmd())

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
