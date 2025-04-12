package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "wallet-service",
		Short: "A gRPC wallet service",
		Long:  `A gRPC wallet service that handles wallet operations in the microservice architecture`,
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
