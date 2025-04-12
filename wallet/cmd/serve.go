package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"wallet-service/internal/config"
	"wallet-service/internal/database"
	"wallet-service/internal/domain/repositories"
	"wallet-service/internal/domain/services"
	"wallet-service/internal/jwt"
	"wallet-service/kafka"
	pb "wallet-service/proto/gen"

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

	trxConsumer := kafka.NewConsumer(dbConn, "deposit_initiated", "deposit_completed")
	defer trxConsumer.Close()

	go trxConsumer.Consume(context.Background())

	// Create dependencies
	walletRepo := repositories.NewPostgresWalletRepository(dbConn)
	jwtUtil := jwt.NewJWTUtil(cfg.JWTKey)
	walletSvc := services.NewWalletService(walletRepo, jwtUtil)

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
