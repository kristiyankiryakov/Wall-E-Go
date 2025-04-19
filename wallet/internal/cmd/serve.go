package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"wallet/internal/config"
	"wallet/internal/consumers"
	"wallet/internal/database"
	"wallet/internal/domain/repositories"
	"wallet/internal/domain/services"
	"wallet/internal/jwt"
	"wallet/internal/producers"
	pb "wallet/proto/gen"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewServeCmd creates and returns the serve command
func NewServeCmd() *cobra.Command {
	var walletPort string

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the wallet service gRPC server",
		Run: func(cmd *cobra.Command, args []string) {
			// Create explicit dependencies for the server
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			if walletPort != "" {
				cfg.WalletPort = walletPort
			}

			serve(cfg)
		},
	}

	// Use flag with default from env
	serveCmd.Flags().StringVarP(&walletPort, "port", "p", os.Getenv("WALLET_PORT"), "Port to run the server on")

	return serveCmd
}

// serve starts the gRPC server with the provided configuration
func serve(cfg *config.Config) {
	// Set environment variables for other components to use
	if err := os.Setenv("DSN", cfg.DSN); err != nil {
		log.Fatalf("Failed to set DSN: %v", err)
	}
	if err := os.Setenv("JWT_KEY", cfg.JWTKey); err != nil {
		log.Fatalf("Failed to set JWT_KEY: %v", err)
	}

	log.Printf("Starting server on :%s...", cfg.WalletPort)

	dbConn := database.ConnectToDB()
	if dbConn == nil {
		log.Panic("Can't connect to Postgres!")
	}
	defer dbConn.Close()

	// Create dependencies
	walletRepo := repositories.NewPostgresWalletRepository(dbConn)
	jwtUtil := jwt.NewJWTUtil(cfg.JWTKey)
	walletSvc := services.NewWalletService(walletRepo, jwtUtil)

	// Initialize producers.
	depositProducer := producers.NewDepositCompletedProducer("localhost:9092", "deposit_completed", 100, 20*time.Millisecond)
	notifyProducer := producers.NewNotificationProducer("localhost:9092", "notification", 100, 20*time.Millisecond)

	// Initialize consumer with config.
	consumerCfg := &consumers.Config{
		Brokers:        []string{"localhost:9092"},
		Topic:          "deposit_initiated",
		GroupID:        "wallet-group",
		BatchSize:      100,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 1 * time.Second,
	}

	consumer := consumers.NewConsumer(dbConn, consumerCfg, depositProducer, notifyProducer)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		consumer.Consume(ctx)
	}()
	defer cancel()
	defer consumer.Close()
	defer depositProducer.Close()
	defer notifyProducer.Close()

	// Set up gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.WalletPort))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, walletSvc)

	log.Printf("Wallet service running on :%s", cfg.WalletPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
