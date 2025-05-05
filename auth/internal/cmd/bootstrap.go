package cmd

import (
	"auth/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
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

	return pool, nil
}

func newLogger(cfg config.Log) *logrus.Logger {
	log := logrus.New()

	// Always write to stdout if enabled:
	writers := []io.Writer{}
	if cfg.StdoutEnabled {
		writers = append(writers, os.Stdout)
	}
	if cfg.FilePath != "" {
		writers = append(writers, &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		})
	}
	log.SetOutput(io.MultiWriter(writers...))

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	switch cfg.Level {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	return log.WithField("service", "auth").Logger
}
