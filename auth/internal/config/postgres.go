package config

import "time"

type Postgres struct {
	Host               string        `default:"localhost" envconfig:"POSTGRES_HOST"`
	Port               int           `default:"5433" envconfig:"POSTGRES_PORT"`
	Username           string        `default:"user" envconfig:"POSTGRES_USERNAME"`
	Password           string        `default:"password" envconfig:"POSTGRES_PASSWORD"`
	Database           string        `default:"auth" envconfig:"POSTGRES_DATABASE"`
	Schema             string        `default:"public" envconfig:"POSTGRES_SCHEMA"`
	SSLMode            string        `default:"disable" envconfig:"POSTGRES_SSL_MODE"`
	ConnectTimeout     string        `default:"5s" envconfig:"POSTGRES_CONNECT_TIMEOUT"`
	MaxOpenConnections int           `default:"10" envconfig:"POSTGRES_MAX_OPEN_CONNECTIONS"`
	MaxIdleConnections int           `default:"10" envconfig:"POSTGRES_MAX_IDLE_CONNECTIONS"`
	MaxConnLifetime    time.Duration `default:"1h" envconfig:"POSTGRES_MAX_CONN_LIFETIME"`
	MaxConnIdleTime    time.Duration `default:"30m" envconfig:"POSTGRES_MAX_CONN_IDLE_TIME"`
}
