package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	rootCmd.Flags().String("config", "", "Path to the config file (eg. config.yaml)")
	_ = viper.BindPFlag("config", rootCmd.Flags().Lookup("config"))

	return rootCmd
}

func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
