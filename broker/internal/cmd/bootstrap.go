package cmd

import (
	"broker/internal/config"
	"broker/internal/handlers"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

type AppHandlers struct {
	Auth        handlers.AuthHandler
	Wallet      handlers.WalletHandler
	Transaction handlers.TransactionHandler
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

	return log.WithField("service", "broker").Logger
}
