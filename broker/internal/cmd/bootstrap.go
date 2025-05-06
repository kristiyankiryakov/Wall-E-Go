package cmd

import (
	"broker/internal/config"
	"broker/internal/handlers"
	"broker/internal/middleware"
	"github.com/gin-gonic/gin"
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

type AppMiddleware struct {
	Auth middleware.AuthMiddleware
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

// SetupRouter configures all routes and middleware for the given router
func setupRouter(
	router *gin.Engine,
	handlers *AppHandlers,
	middleware *AppMiddleware,
) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API versioning
	v1 := router.Group("/api/v1")
	{
		//Health checks
		v1.GET("/health/wallet", handlers.Wallet.HealthCheck)

		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Auth.Authenticate)
			auth.POST("/register", handlers.Auth.Register)
		}

		// Protected routes
		protected := v1.Group("")

		protected.Use(middleware.Auth.AuthenticateUser(), middleware.Auth.AppendUserIDToGrpcContext())
		{
			protected.POST("/wallet", handlers.Wallet.CreateWallet)

			wallet := protected.Group("/wallet")
			{
				wallet.GET("", handlers.Wallet.ViewBalance)
			}

			// Wallet owner routes
			ownerProtected := protected.Group("/transaction")
			ownerProtected.Use(middleware.Auth.AuthenticateWalletOwner())
			{
				ownerProtected.PUT("/deposit", handlers.Transaction.Deposit)
			}
		}
	}
}
