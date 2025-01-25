package main

import (
	"database/sql"
	"log"
	"wall-e-go/auth-service/config"
	"wall-e-go/auth-service/internal/handlers"
	"wall-e-go/auth-service/internal/middleware"
	"wall-e-go/auth-service/internal/repository"
	"wall-e-go/auth-service/internal/services"
	jwt "wall-e-go/auth-service/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load config
	cfg := config.LoadConfig()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to database
	dsn := "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Set up repositories services and handlers

	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(*userRepo, jwt.JWTUtil{})
	authHandler := handlers.NewAuthHandler(authService)

	router := gin.Default()

	// Use error handler middleware
	router.Use(middleware.ErrorHandler())

	router.POST("/register", authHandler.Register)

	// Start server
	log.Println("Starting server on :8080...")
	router.Run(":8080")
}
