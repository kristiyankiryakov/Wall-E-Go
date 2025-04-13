package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"notification/internal/consumer"
	"notification/logger"
)

func NewServeCmd() *cobra.Command {
	serveCmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the transaction service gRPC server",
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.NewLogger()

			log.Info("Starting consumer...")

			trxCompletedConsumer := consumer.NewConsumer("deposit_completed", "notification")
			defer trxCompletedConsumer.Close()

			go trxCompletedConsumer.Consume(context.Background())

		},
	}

	return serveCmdInstance
}
