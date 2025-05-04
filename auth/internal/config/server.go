package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type ServerCfg struct {
	// ListenPort is the port where the server listens for incoming requests.
	ListenPort string `default:"50051" envconfig:"LISTEN_PORT"`

	JWTSecret string `default:"change-me-in-prod" envconfig:"JWT_SECRET"`

	Postgres Postgres
}

func NewServerConfig() (*ServerCfg, error) {
	var cfg ServerCfg

	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process env variables: %w", err)
	}

	return &cfg, nil
}
