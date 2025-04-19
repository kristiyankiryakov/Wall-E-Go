package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"notification/internal/channel/mail"
	"notification/internal/config"
	"notification/internal/consumer"
	"notification/internal/service"
	"notification/logger"
)

func NewServeCmd() *cobra.Command {
	serveCmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the notification service",
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.NewLogger()
			ctx := context.Background()

			cfg := config.LoadConfig()

			// Initialize notification service
			notificationSvc := service.NewNotificationService()

			// Register mail channel
			mailSender := mail.NewMail(cfg.Mail)
			err := notificationSvc.AddChannel("email", mailSender)
			if err != nil {
				log.Fatalf("Failed to register mail channel: %v", err)
			}

			// Initialize consumer with dependency injection
			log.Info("Starting consumer...")
			notificationConsumer := consumer.NewConsumer(
				cfg.Kafka,
				notificationSvc,
			)
			defer notificationConsumer.Close()

			notificationConsumer.Consume(ctx)
		},
	}

	return serveCmdInstance
}
