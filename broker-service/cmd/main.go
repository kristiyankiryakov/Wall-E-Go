package main

import (
	"broker-service/internal/authclient"
	"broker-service/internal/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func main() {
	authClient, err := authclient.NewAuthClient("auth:50051")
	if err != nil {
		log.Fatal(err)
	}

	handler := handlers.NewBrokerHandler(authClient)
	router := handlers.SetupRouter(handler)

	log.Println("Broker service running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
