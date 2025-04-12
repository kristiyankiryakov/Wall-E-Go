package config

import (
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	DSN        string `mapstructure:"DSN"`
	WalletPort string `mapstructure:"WALLET_PORT"`
	JWTKey     string `mapstructure:"JWT_KEY"`
}

func LoadConfig() (*Config, error) {
	// First try env variables
	config := &Config{
		DSN:        os.Getenv("DSN"),
		WalletPort: os.Getenv("WALLET_PORT"),
		JWTKey:     os.Getenv("JWT_KEY"),
	}

	// If DSN is empty, try loading from config file
	if config.DSN == "" {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")

		if err := viper.ReadInConfig(); err != nil {
			// Config file not found, use defaults
			config.DSN = "host=localhost port=5434 user=user password=password dbname=wallet_db sslmode=disable"
			if config.WalletPort == "" {
				config.WalletPort = "50052"
			}
		} else {
			if err := viper.Unmarshal(config); err != nil {
				return nil, err
			}
		}
	}

	return config, nil
}
