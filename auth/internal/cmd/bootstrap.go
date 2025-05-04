package cmd

import (
	"auth/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newPostgresPool(ctx context.Context, cfg config.Postgres) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	pgxPoolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	pgxPoolCfg.MaxConns = int32(cfg.MaxOpenConnections)
	pgxPoolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxPoolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	pgxPoolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, pgxPoolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping pgx pool: %w", err)
	}

	fmt.Println("Successfully connected to Postgres database")

	return pool, nil
}
