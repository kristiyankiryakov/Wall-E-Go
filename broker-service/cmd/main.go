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

	handler := handlers.NewBrokerHandler(authClient, walletClient)
	router := handlers.SetupRouter(handler)

	log.Println("Broker service running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
