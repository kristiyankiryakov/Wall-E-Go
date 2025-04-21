package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture simulates a capture process via fake SFTP serve",
		Long:  `Capture simulates a capture process via fake SFTP serve`,
	}
	rootCmd.AddCommand(NewCaptureCmd())

	return rootCmd
}

func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
