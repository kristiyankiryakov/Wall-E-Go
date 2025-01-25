package main

import (
	"database/sql"
	"log"
	"wall-e-go/auth-service/config"
	"wall-e-go/auth-service/internal/handlers"
	"wall-e-go/auth-service/internal/middleware"
	"wall-e-go/auth-service/internal/repository"
	router "wall-e-go/auth-service/internal/routers"
	"wall-e-go/auth-service/internal/services"
	jwt "wall-e-go/auth-service/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
	Router *gin.Engine
}

func main() {

	app := InitializeApp()

	// Start server
	log.Println("Starting server on :8080...")
	app.Router.Run(":8080")
}

func InitializeApp() *App {
	cfg := config.LoadConfig()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db := initializeDatabase(cfg)

	r := initializeRouter(db)

	return &App{
		Config: cfg,
		DB:     db,
		Router: r,
	}
}

func initializeDatabase(cfg *config.Config) *sql.DB {
	dsn := "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

// sets up the router and handlers

func initializeRouter(db *sql.DB) *gin.Engine {
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(*userRepo, jwt.JWTUtil{})
	authHandler := handlers.NewAuthHandler(authService)

	r := gin.Default()

	r.Use(middleware.ErrorHandler())

	routes := router.NewRouter(authHandler)
	routes.RegisterRoutes(r)

	return r
}
