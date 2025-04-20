package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
)

// Config holds application configuration
type Config struct {
	AuthHost        string
	WalletHost      string
	TransactionHost string
	ServerPort      string
	JwtKey          string
}

func NewConfig() *Config {
	// Load .env file (ignore if not found)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values or system environment variables")
	}

	// Set default values
	viper.SetDefault("auth_host", "localhost:50051")
	viper.SetDefault("wallet_host", "localhost:50052")
	viper.SetDefault("transaction_host", "localhost:50053")
	viper.SetDefault("server_port", "8080")
	viper.SetDefault("jwt_key", "default-jwt-key")

	viper.SetEnvPrefix("BROKER")
	viper.AutomaticEnv() // maps BROKER_AUTH_HOST to auth_host, etc.

	// Looks for a config file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/broker/") // Production path
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return &Config{
		AuthHost:        viper.GetString("auth_host"),
		WalletHost:      viper.GetString("wallet_host"),
		TransactionHost: viper.GetString("transaction_host"),
		ServerPort:      viper.GetString("server_port"),
		JwtKey:          viper.GetString("jwt_key"),
	}
}
