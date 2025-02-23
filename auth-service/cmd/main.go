package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"wall-e-go/internal/data"
	"wall-e-go/internal/jwt"
	"wall-e-go/internal/service"
	pb "wall-e-go/proto"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"google.golang.org/grpc"
)

const WEB_PORT = "50001"

func main() {
	log.Printf("Starting server on :%s...", WEB_PORT)

	dbConn := connectToDB()
	if dbConn == nil {
		log.Panic("Can't connect to Postgres!")
	}
	defer dbConn.Close()

	userRepo := data.NewPostgresUserRepository(dbConn)
	jwtUtil := jwt.NewJWTUtil(os.Getenv("JWT_KEY"))
	authSvc := service.NewAuthService(jwtUtil, userRepo)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", WEB_PORT))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, authSvc)
	log.Printf("Auth service running on :%s", WEB_PORT)
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
