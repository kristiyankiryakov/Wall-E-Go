package cmd

import (
	"broker-service/internal/clients"
	"broker-service/internal/handlers"
	"broker-service/internal/middleware"
	"broker-service/internal/routes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"log"

	"broker-service/internal/config"
)

// Server represents the API gateway server
type Server struct {
	config *config.Config
	router *gin.Engine
}

// NewServer creates a new server instance with dependencies
func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
		router: gin.Default(),
	}
}

// NewServeCmd creates the serve command
func NewServeCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the API gateway server",
		RunE: func(cmd *cobra.Command, args []string) error {
			appConfig := config.NewConfig()
			server := NewServer(appConfig)
			err := server.Start()
			if err != nil {
				log.Fatalf("Failed to start server: %v", err)
				return err
			}

			return nil
		},
	}

	return cmdInstance
}

// Start initializes and starts the API gateway
func (s *Server) Start() error {
	// Initialize config
	cfg := config.NewConfig()

	// Initialize clients
	authClient, err := clients.NewAuthClient(s.config.AuthHost)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	walletClient, err := clients.NewWalletClient(s.config.WalletHost)
	if err != nil {
		return fmt.Errorf("failed to create wallet client: %w", err)
	}

	transactionClient, err := clients.NewTransactionClient(s.config.TransactionHost)
	if err != nil {
		return fmt.Errorf("failed to create transaction client: %w", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authClient)
	walletHandler := handlers.NewWalletHandler(walletClient)
	transactionHandler := handlers.NewTransactionHandler(transactionClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg, walletClient)

	// Set up routes
	routes.SetupRouter(
		s.router,
		authHandler,
		walletHandler,
		transactionHandler,
		walletClient,
		authMiddleware,
	)

	// Start server
	log.Printf("Broker service running on :%s", s.config.ServerPort)
	if err := s.router.Run(":" + s.config.ServerPort); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
