package cmd

import (
	"auth/internal/auth"
	"auth/internal/config"
	"auth/internal/jwt"
	"auth/internal/user"
	pb "auth/proto/gen"
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
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

			log := newLogger(cfg.Log)

			pgPool, err := newPostgresPool(ctx, cfg.Postgres)
			if err != nil {
				return fmt.Errorf("failed to create postgres pool: %w", err)
			}

			log.Info("Successfully connected to postgres")

			userRepo := user.NewPostgresUserRepository(pgPool, log)
			jwtUtil := jwt.NewJWTUtil(cfg.JWTSecret)
			authSvc := auth.NewAuthService(jwtUtil, userRepo, log)

			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.ListenPort))
			if err != nil {
				return fmt.Errorf("failed to listen: %w", err)
			}

			log.Infof("gRPC server listening on port %s", cfg.ListenPort)
			s := grpc.NewServer()
			pb.RegisterAuthServiceServer(s, authSvc)
			if err := s.Serve(lis); err != nil {
				return fmt.Errorf("failed to serve: %w", err)
			}

			return nil
		},
	}

	return cmdInstance
}
