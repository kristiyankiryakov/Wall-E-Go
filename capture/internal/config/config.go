package config

import "github.com/spf13/viper"

type Config struct {
	DatabaseUrl string
}

func NewConfig() *Config {
	viper.SetDefault("database_url", "postgres://user:password@localhost:5435/transaction_db?sslmode=disable")

	viper.SetEnvPrefix("CAPTURE")
	viper.AutomaticEnv()

	return &Config{
		DatabaseUrl: viper.GetString("database_url"),
	}
}
