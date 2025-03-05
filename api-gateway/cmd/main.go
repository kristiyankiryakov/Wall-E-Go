package main

import (
	"broker-service/internal/clients"
	"broker-service/internal/handlers"
	"broker-service/internal/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func main() {
	authClient, err := clients.NewAuthClient(os.Getenv("AUTH_HOST"))
	if err != nil {
		log.Fatal(err)
	}
	walletClient, err := clients.NewWalletClient(os.Getenv("WALLET_HOST"))
	if err != nil {
		log.Fatal(err)
	}
	transactionClient, err := clients.NewTransactionClient(os.Getenv("TRANSACTION_HOST"))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authClient)
	walletHandler := handlers.NewWalletHandler(walletClient)
	transactionHandler := handlers.NewTransactionHandler(transactionClient)

	r := gin.Default()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Authenticate)
		authGroup.POST("/register", authHandler.Register)
	}

	{
		r.POST("/create", middleware.AuthenticateUser(), middleware.AppendUserIDToGrpcContext(), walletHandler.CreateWallet)
	}

	protectedWalletGroup := r.Group("/wallet")
	protectedWalletGroup.Use(middleware.AuthenticateUser(), middleware.AppendUserIDToGrpcContext())
	{
		protectedWalletGroup.GET("/view", walletHandler.ViewBalance)
	}

	txGroup := r.Group("/transaction")
	txGroup.Use(middleware.AuthenticateWalletOwner(walletClient), middleware.AppendUserIDToGrpcContext())
	{
		txGroup.PUT("/deposit", transactionHandler.Deposit)
	}

	log.Println("Broker service running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
