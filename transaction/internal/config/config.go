package config

import "github.com/spf13/viper"

type Config struct {
	TRANSACTION_GRPC_HOST string
	KAFKA_HOST            string
	GRPC_PORT             string
	DSN                   string
}

func NewConfig() *Config {
	viper.SetDefault("transaction_grpc_host", "localhost:50053")
	viper.SetDefault("kafka_host", "localhost:9092")
	viper.SetDefault("grpc_port", "50053")
	viper.SetDefault("dsn", "host=localhost port=5435 user=user password=password dbname=transaction_db sslmode=disable timezone=UTC connect_timeout=5")

	viper.SetEnvPrefix("TRANSACTION")
	viper.AutomaticEnv() // maps TRANSACTION_GRPC_HOST to TRANSACTION_GRPC_HOST, etc.

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/transaction/") // Production path
	if err := viper.ReadInConfig(); err == nil {
		println("Using config file:", viper.ConfigFileUsed())
	}

	return &Config{
		TRANSACTION_GRPC_HOST: viper.GetString("transaction_grpc_host"),
		KAFKA_HOST:            viper.GetString("kafka_host"),
		GRPC_PORT:             viper.GetString("grpc_port"),
		DSN:                   viper.GetString("dsn"),
	}
}
