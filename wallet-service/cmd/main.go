package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"wallet-service/internal/data"
	"wallet-service/internal/jwt"
	"wallet-service/internal/service"
	pb "wallet-service/proto"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"google.golang.org/grpc"
)

const gRPC_PORT = "50002"

func main() {
	log.Printf("Starting server on :%s...", gRPC_PORT)

	dbConn := connectToDB()
	if dbConn == nil {
		log.Panic("Can't connect to Postgres!")
	}
	defer dbConn.Close()

	walletRepo := data.NewPostgresWalletRepository(dbConn)
	jwtUtil := jwt.NewJWTUtil(os.Getenv("JWT_KEY"))
	walletSvc := service.NewWalletService(walletRepo, jwtUtil)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRPC_PORT))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, walletSvc)

	log.Printf("Wallet service running on :%s", gRPC_PORT)
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
