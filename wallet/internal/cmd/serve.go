package cmd

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"net"
	"time"
	"wallet/internal/config"
	"wallet/internal/consumers"
	"wallet/internal/jwt"
	"wallet/internal/producers"
	"wallet/internal/wallet"
	pb "wallet/proto/gen"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewServeCmd creates and returns the serve command
func NewServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the wallet service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load .env file
			if err := godotenv.Load(); err != nil {
				return fmt.Errorf("failed to load .env file: %w", err)
			}

			cfg, err := config.NewServerConfig()
			if err != nil {
				return fmt.Errorf("failed to create runtime config: %w", err)
			}

			log := newLogger(cfg.Log)

			pgPool, err := newPostgresPool(ctx, cfg.Postgres)
			if err != nil {
				return fmt.Errorf("failed to create postgres pool: %w", err)
			}

			log.Info("Successfully connected to postgres")

			// Create dependencies
			walletRepo := wallet.NewPostgresWalletRepository(pgPool)
			jwtUtil := jwt.NewJWTUtil(cfg.JWTSecret)
			walletSvc := wallet.NewWalletService(walletRepo, jwtUtil, log)

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

			consumer := consumers.NewConsumer(pgPool, consumerCfg, depositProducer, notifyProducer)
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				consumer.Consume(ctx)
			}()
			defer cancel()
			defer consumer.Close()
			defer depositProducer.Close()
			defer notifyProducer.Close()

			// Set up gRPC server
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.ListenPort))
			if err != nil {
				log.Fatal(err)
			}

			s := grpc.NewServer()
			pb.RegisterWalletServiceServer(s, walletSvc)

			log.Printf("Wallet service running on :%s", cfg.ListenPort)
			if err := s.Serve(lis); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	return serveCmd
}
