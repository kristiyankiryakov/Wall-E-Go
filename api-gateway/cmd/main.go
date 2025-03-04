package main

import (
	"broker-service/internal/clients"
	"broker-service/internal/handlers"
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

	walletGroup := r.Group("/wallet")
	{
		walletGroup.POST("/create", walletHandler.CreateWallet)
		walletGroup.GET("/view", walletHandler.ViewBalance)
	}
	txGroup := r.Group("/transaction")
	{
		txGroup.PUT("/deposit", transactionHandler.Deposit)
	}

	log.Println("Broker service running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
