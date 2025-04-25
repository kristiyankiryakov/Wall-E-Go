package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	AuthPort string
	JwtKey   string
	DSN      string
}

func NewConfig() *Config {
	// Load .env file (ignore if not found)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values or system environment variables")
	}

	// Set default values
	viper.SetDefault("auth_port", "50051")
	viper.SetDefault("jwt_key", "default-jwt-key")
	viper.SetDefault("dsn", "host=auth-db port=5432 user=user password=password dbname=auth_service sslmode=disable timezone=UTC connect_timeout=5")

	viper.AutomaticEnv()

	// Looks for a config file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/auth/") // Production path
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return &Config{
		AuthPort: viper.GetString("auth_port"),
		JwtKey:   viper.GetString("jwt_key"),
		DSN:      viper.GetString("dsn"),
	}
}
