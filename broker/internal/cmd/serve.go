package cmd

import (
	"broker/internal/clients"
	"broker/internal/config"
	"broker/internal/handlers"
	"broker/internal/middleware"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// NewServeCmd creates the serve command
func NewServeCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the API gateway server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := godotenv.Load(); err != nil {
				return fmt.Errorf("failed to load .env file: %w", err)
			}

			cfg, err := config.NewServerConfig()
			if err != nil {
				return fmt.Errorf("failed to create runtime config: %w", err)
			}

			log := newLogger(cfg.Log)

			// Initialize clients
			authClient, err := clients.NewAuthClient(cfg.AuthHost, log)
			if err != nil {
				log.WithError(err).Error("Failed to create auth client")
				return fmt.Errorf("failed to create auth client: %w", err)
			}

			walletClient, err := clients.NewWalletClient(cfg.WalletHost, log)
			if err != nil {
				log.WithError(err).Error("Failed to create wallet client")
				return fmt.Errorf("failed to create wallet client: %w", err)
			}

			transactionClient, err := clients.NewTransactionClient(cfg.TransactionHost, log)
			if err != nil {
				log.WithError(err).Error("Failed to create transaction client")
				return fmt.Errorf("failed to create transaction client: %w", err)
			}

			appHandlers := &AppHandlers{
				Auth:        handlers.NewAuthHandler(authClient),
				Wallet:      handlers.NewWalletHandler(walletClient),
				Transaction: handlers.NewTransactionHandler(transactionClient),
			}

			appMiddleware := &AppMiddleware{
				Auth: middleware.NewAuthMiddleware(cfg, walletClient, log),
			}

			// Initialize Gin router
			router := gin.Default()
			setupRouter(router, appHandlers, appMiddleware)

			log.Info("Starting server on port " + cfg.ListenPort)
			if err := router.Run(":" + cfg.ListenPort); err != nil {
				return fmt.Errorf("failed to run server: %w", err)
			}

			if err != nil {
				log.WithError(err).Error("Failed to start server")
				return fmt.Errorf("failed to start server: %w", err)
			}

			return nil
		},
	}

	return cmdInstance
}
