package routes

import (
	"broker-service/internal/clients"
	"broker-service/internal/handlers"
	"broker-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware for the given router
func SetupRouter(
	router *gin.Engine,
	authHandler handlers.AuthHandler,
	walletHandler handlers.WalletHandler,
	transactionHandler handlers.TransactionHandler,
	walletClient *clients.WalletClient,
	authMiddleware middleware.AuthMiddleware,
) {
	// Auth routes
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Authenticate)
		authGroup.POST("/register", authHandler.Register)
	}

	// Wallet routes
	router.POST("/create", authMiddleware.AuthenticateUser(), authMiddleware.AppendUserIDToGrpcContext(), walletHandler.CreateWallet)

	protectedWalletGroup := router.Group("/wallet")
	protectedWalletGroup.Use(authMiddleware.AuthenticateUser(), authMiddleware.AppendUserIDToGrpcContext())
	{
		protectedWalletGroup.GET("/view", walletHandler.ViewBalance)
	}

	// Transaction routes
	txGroup := router.Group("/transaction")
	txGroup.Use(authMiddleware.AuthenticateWalletOwner(walletClient), authMiddleware.AppendUserIDToGrpcContext())
	{
		txGroup.PUT("/deposit", transactionHandler.Deposit)
	}
}
