package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// NewRootCmd creates and returns the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "transaction-service",
		Short: "A gRPC transaction service",
		Long:  `A gRPC transaction service that handles transaction operations in the microservice architecture`,
	}

	// Add commands
	rootCmd.AddCommand(NewServeCmd())

	return rootCmd
}

// Execute runs the root command
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
