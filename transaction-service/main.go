package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"transaction-service/internal/data"
	"transaction-service/internal/service"
	"transaction-service/kafka"
	pb "transaction-service/proto/gen"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"google.golang.org/grpc"
)

var gRPC_PORT = os.Getenv("TRANSACTION_PORT")

const (
	DEPOSIT_INITIATED string = "deposit_initiated"
	DEPOSIT_COMPLETED string = "deposit_completed"
)

func main() {
	log.Printf("Starting server on :%s...", gRPC_PORT)

	dbConn := connectToDB()
	if dbConn == nil {
		log.Panic("Can't connect to Postgres!")
	}
	defer dbConn.Close()

	trxProducer := kafka.NewProducer(DEPOSIT_INITIATED)
	defer trxProducer.Close()

	trxConsumer := kafka.NewConsumer(dbConn, DEPOSIT_COMPLETED)
	defer trxConsumer.Close()

	//runs in a goroutine
	go trxConsumer.Consume(context.Background())

	tsxRepo := data.NewPostgresTransactionRepository(dbConn)
	tsxSvc := service.NewTransactionService(tsxRepo, trxProducer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRPC_PORT))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterTransactionServiceServer(s, tsxSvc)

	log.Printf("Transaction service running on port :%s", gRPC_PORT)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	db.SetMaxOpenConns(25)

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	counts := 0

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println(err)
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("Connected to Postgres")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}
