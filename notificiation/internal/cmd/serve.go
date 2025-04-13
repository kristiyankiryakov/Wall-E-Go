package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"notification/internal/channel/mail"
	"notification/internal/config"
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
			mailSender := mail.NewMail(&config.MailConfig{
				SMTPHost: "localhost",
				SMTPPort: "1025",
				Auth:     nil,
			})
			trxCompletedConsumer := consumer.NewConsumer("deposit_completed", "notification", mailSender)
			defer trxCompletedConsumer.Close()

			trxCompletedConsumer.Consume(context.Background())
		},
	}

	return serveCmdInstance
}
