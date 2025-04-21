package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCaptureCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "capture",
		Short: "Capture simulates a capture process via fake SFTP serve",
		Long:  `Capture simulates a capture process via fake SFTP serve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ExecuteCaptureAuthTransactions()

			return nil
		},
	}
	cmdInstance.Flags().String("config", "", "Path to the config file (eg. config.yaml)")
	_ = viper.BindPFlag("config", cmdInstance.Flags().Lookup("config"))
	return cmdInstance
}

func ExecuteCaptureAuthTransactions() {

}
