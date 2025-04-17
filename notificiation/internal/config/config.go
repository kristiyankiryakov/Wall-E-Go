package config

import "net/smtp"

type Mail struct {
	SMTPHost string
	SMTPPort string
	Auth     smtp.Auth
}
type Kafka struct {
	Brokers []string
	Topic   string
	GroupID string
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
			Brokers: []string{"localhost:9092"},
			Topic:   "deposit_completed",
			GroupID: "notification-group",
		},
	}
}
