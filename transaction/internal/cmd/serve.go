package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"log"
	"net"
	"transaction/internal/config"
	"transaction/internal/consumer"
	"transaction/internal/database"
	"transaction/internal/domain/repositories"
	"transaction/internal/domain/services"
	"transaction/internal/producer"
	pb "transaction/proto/gen"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	DEPOSIT_INITIATED string = "deposit_initiated"
	DEPOSIT_COMPLETED string = "deposit_completed"
)

func NewServeCmd() *cobra.Command {
	serveCmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the transaction service gRPC server",
		Run: func(cmd *cobra.Command, args []string) {
			// Create explicit dependencies for the server
			cfg := config.NewConfig()

			log.Printf("Starting server on :%s...", cfg.GRPC_PORT)

			dbConn := database.ConnectToDB(cfg.DSN)
			if dbConn == nil {
				log.Panic("Can't connect to Postgres!")
			}
			defer func(dbConn *sql.DB) {
				err := dbConn.Close()
				if err != nil {
					log.Fatalf("Failed to close database connection: %v", err)
				} else {
					log.Println("Database connection closed successfully.")
				}
			}(dbConn)
			tsxRepo := repositories.NewPostgresTransactionRepository(dbConn)

			trxProducer := producer.NewProducer(DEPOSIT_INITIATED)
			defer trxProducer.Close()

			trxConsumer := consumer.NewConsumer(DEPOSIT_COMPLETED, tsxRepo)
			defer trxConsumer.Close()

			go trxConsumer.Consume(context.Background())

			tsxSvc := services.NewTransactionService(tsxRepo, trxProducer)

			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC_PORT))
			if err != nil {
				log.Fatal(err)
			}

			s := grpc.NewServer()
			pb.RegisterTransactionServiceServer(s, tsxSvc)

			log.Printf("Transaction service running on port :%s", cfg.GRPC_PORT)
			if err := s.Serve(lis); err != nil {
				log.Fatal(err)
			}
		},
	}

	return serveCmdInstance
}
