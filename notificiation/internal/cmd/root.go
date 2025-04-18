package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// NewRootCmd creates and returns the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "notification-service",
		Short: "A gRPC notification service",
		Long:  `A gRPC notification service`,
	}

	rootCmd.AddCommand(NewServeCmd())

	return rootCmd
}

func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
