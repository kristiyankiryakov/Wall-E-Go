package main

import (
	"broker-service/internal/clients"
	"broker-service/internal/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func main() {
	authClient, err := clients.NewAuthClient("auth:50001")
	if err != nil {
		log.Fatal(err)
	}
	walletClient, err := clients.NewWalletClient("wallet:50002")
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
