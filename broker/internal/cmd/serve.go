package cmd

import (
	"broker/internal/clients"
	"broker/internal/config"
	"broker/internal/handlers"
	"broker/internal/middlewares"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
)

// NewServeCmd creates the serve command
func NewServeCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the API gateway server",
		RunE: func(cmd *cobra.Command, args []string) error {
			//ctx := context.Background()

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

			//transactionClient, err := clients.NewTransactionClient(cfg.TransactionHost, log)
			//if err != nil {
			//	log.WithError(err).Error("Failed to create transaction client")
			//	return fmt.Errorf("failed to create transaction client: %w", err)
			//}

			// Initialize handlers
			authHandler := handlers.NewAuthHandler(authClient)
			walletHandler := handlers.NewWalletHandler(walletClient)
			//transactionHandler := handlers.NewTransactionHandler(transactionClient)

			// Initialize router
			router := chi.NewRouter()

			router.Route("/api/v1", func(v1 chi.Router) {
				v1.Use(middlewares.RequestID)
				v1.Use(middlewares.Tracer)
				v1.Use(middlewares.Logger(log))

				v1.Use(cors.Handler(cors.Options{
					AllowedOrigins: []string{"*"},
					AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				}))

				// Public routes
				v1.Route("/auth", func(auth chi.Router) {
					auth.Post("/login", authHandler.Authenticate)
					auth.Post("/register", authHandler.Register)
				})
				v1.Route("/health", func(health chi.Router) {
					health.Get("/wallet", walletHandler.HealthCheck)
				})

				v1.Route("/", func(protected chi.Router) {
					protected.Use(middlewares.Authenticate(cfg.JWTSecret, log))

					protected.Post("/wallet", walletHandler.CreateWallet)
					protected.Get("/wallet", walletHandler.ViewBalance)
				})
			})

			host := fmt.Sprintf(cfg.ListenHost + ":" + cfg.ListenPort)
			log.Info("Starting server on addr: ", host)
			err = http.ListenAndServe(host, router)
			if err != nil {
				log.WithError(err).Error("Failed to start server")
				return fmt.Errorf("failed to start server: %w", err)
			}

			return nil
		},
	}

	return cmdInstance
}
