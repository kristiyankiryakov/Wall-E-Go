package config

type Config struct {
	TRANSACTION_GRPC_HOST string
	KAFKA_HOST            string
	GRPC_PORT             string
	DSN                   string
}

func NewConfig() *Config {
	TRANSACTION_GRPC_HOST := "localhost:50053"
	KAFKA_HOST := "localhost:9092"
	GRPC_PORT := "50053"
	DSN := "host=localhost port=5435 user=user password=password dbname=transaction_db sslmode=disable timezone=UTC connect_timeout=5"

	return &Config{
		TRANSACTION_GRPC_HOST,
		KAFKA_HOST,
		GRPC_PORT,
		DSN,
	}
}
