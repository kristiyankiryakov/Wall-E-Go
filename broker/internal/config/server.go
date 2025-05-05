package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type ServerCfg struct {
	// ListenPort is the port where the server listens for incoming requests.
	ListenPort string `default:"8080" envconfig:"LISTEN_PORT"`

	// JWTSecret is the secret key used for signing JWT tokens.
	JWTSecret string `default:"change-me-in-prod" envconfig:"JWT_SECRET"`

	AuthHost        string `default:"localhost:50051" envconfig:"AUTH_HOST"`
	WalletHost      string `default:"localhost:50052" envconfig:"WALLET_HOST"`
	TransactionHost string `default:"localhost:50053" envconfig:"TRANSACTION_HOST"`

	Log Log
}

func NewServerConfig() (*ServerCfg, error) {
	var cfg ServerCfg

	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process env variables: %w", err)
	}

	return &cfg, nil
}
