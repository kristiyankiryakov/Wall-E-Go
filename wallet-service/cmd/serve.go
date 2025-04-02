package cmd

import (
	"context"
	"wallet-service/internal/config"

	"fmt"
	"log"
	"net"
	"os"
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

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the wallet service gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Use flag with default from env
	serveCmd.Flags().StringVarP(&port, "port", "p", os.Getenv("WALLET_PORT"), "Port to run the server on")
}

func serve() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use command-line port if provided, otherwise use config
	if port == "" {
		port = cfg.WalletPort
	}

	// Set environment variables for other components to use
	err = os.Setenv("DSN", cfg.DSN)
	if err != nil {
		log.Fatalf("Failed to set DSN: %v", err)
	}
	err = os.Setenv("JWT_KEY", cfg.JWTKey)
	if err != nil {
		log.Fatalf("Failed to set JWT_KEY")
	}

	log.Printf("Starting server on :%s...", port)

	dbConn := database.ConnectToDB()
	if dbConn == nil {
		log.Panic("Can't connect to Postgres!")
	}
	defer dbConn.Close()

	trxConsumer := kafka.NewConsumer(dbConn, "deposit_initiated", "deposit_completed")
	//runs in a goroutine
	go trxConsumer.Consume(context.Background())

	walletRepo := repositories.NewPostgresWalletRepository(dbConn)
	jwtUtil := jwt.NewJWTUtil(os.Getenv("JWT_KEY"))
	walletSvc := services.NewWalletService(walletRepo, jwtUtil)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, walletSvc)

	log.Printf("Wallet service running on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
