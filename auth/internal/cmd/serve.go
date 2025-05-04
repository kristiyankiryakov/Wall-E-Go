package cmd

import (
	"auth/internal/config"
	"auth/internal/data"
	"auth/internal/jwt"
	"auth/internal/service"
	pb "auth/proto/gen"
	"context"
	"fmt"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"log"
	"net"
)

// NewServeCmd creates the serve command
func NewServeCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "serve",
		Short: "Start the auth-service",
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

			pgPool, err := newPostgresPool(ctx, cfg.Postgres)
			if err != nil {
				return fmt.Errorf("failed to create postgres pool: %w", err)
			}

			userRepo := data.NewPostgresUserRepository(pgPool)
			jwtUtil := jwt.NewJWTUtil(cfg.JWTSecret)
			authSvc := service.NewAuthService(jwtUtil, userRepo)

			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.ListenPort))
			if err != nil {
				return fmt.Errorf("failed to listen: %w", err)
			}

			s := grpc.NewServer()
			pb.RegisterAuthServiceServer(s, authSvc)
			log.Printf("Auth service running on :%s", cfg.ListenPort)
			if err := s.Serve(lis); err != nil {
				return fmt.Errorf("failed to serve: %w", err)
			}

			fmt.Printf("Auth service running on %s\n", cfg.ListenPort)
			return nil
		},
	}

	return cmdInstance
}
