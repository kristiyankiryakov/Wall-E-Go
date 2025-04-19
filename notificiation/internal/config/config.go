package config

import (
	"net/smtp"
	"time"
)

type Mail struct {
	SMTPHost string
	SMTPPort string
	Auth     smtp.Auth
}
type Kafka struct {
	Brokers        []string
	Topic          string
	GroupID        string
	NumWorkers     int
	MinBytes       int
	MaxBytes       int
	CommitInterval time.Duration
	BatchSize      int
	BatchTimeout   time.Duration
}

type Config struct {
	*Mail
	*Kafka
}

func LoadConfig() *Config {
	return &Config{
		&Mail{
			SMTPHost: "localhost",
			SMTPPort: "1025",
			Auth:     nil,
		},
		&Kafka{
			Brokers:        []string{"localhost:9092"},
			Topic:          "notification",
			GroupID:        "notification-group",
			NumWorkers:     5,
			MinBytes:       10e3, // 10KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: 1 * time.Second,
			BatchSize:      100,
			BatchTimeout:   1 * time.Second,
		},
	}
}
