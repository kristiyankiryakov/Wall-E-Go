package cmd

import (
	"auth/internal/config"
	"auth/internal/data"
	"auth/internal/jwt"
	"auth/internal/service"
	pb "auth/proto/gen"
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// NewServeCmd creates the serve command
func NewServeCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the auth-service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.NewConfig()
			if cfg == nil {
				log.Fatal("Failed to load config")
			}

			log.Printf("Starting server on :%s...", cfg.AuthPort)

			dbConn := connectToDB()
			if dbConn == nil {
				log.Panic("Can't connect to Postgres!")
			}
			defer dbConn.Close()

			userRepo := data.NewPostgresUserRepository(dbConn)
			jwtUtil := jwt.NewJWTUtil(cfg.JwtKey)
			authSvc := service.NewAuthService(jwtUtil, userRepo)

			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.AuthPort))
			if err != nil {
				log.Fatal(err)
			}

			s := grpc.NewServer()
			pb.RegisterAuthServiceServer(s, authSvc)
			log.Printf("Auth service running on :%s", cfg.AuthPort)
			if err := s.Serve(lis); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}
	cmdInstance.Flags().String("config", "", "Path to the config file (eg. config.yaml)")
	_ = viper.BindPFlag("config", cmdInstance.Flags().Lookup("config"))

	return cmdInstance
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
