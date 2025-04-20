package config

import (
	"fmt"
	"github.com/spf13/viper"
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
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topic", "notification")
	viper.SetDefault("kafka.group_id", "notification-group")
	viper.SetDefault("kafka.num_workers", 5)
	viper.SetDefault("kafka.min_bytes", 10e3) // 10KB
	viper.SetDefault("kafka.max_bytes", 10e6) // 10MB
	viper.SetDefault("kafka.commit_interval", 1*time.Second)
	viper.SetDefault("kafka.batch_size", 100)
	viper.SetDefault("kafka.batch_timeout", 1*time.Second)

	viper.SetDefault("mail.smtp_host", "localhost")
	viper.SetDefault("mail.smtp_port", "1025")
	viper.SetDefault("mail.auth", nil)

	viper.SetEnvPrefix("NOTIFICATION")
	viper.AutomaticEnv() // maps NOTIFICATION_KAFKA_BROKERS to kafka.brokers, etc.

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/notification/") // Production path
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return &Config{
		&Mail{
			SMTPHost: viper.GetString("mail.smtp_host"),
			SMTPPort: viper.GetString("mail.smtp_port"),
			Auth:     nil,
		},
		&Kafka{
			Brokers:        viper.GetStringSlice("kafka.brokers"),
			Topic:          viper.GetString("kafka.topic"),
			GroupID:        viper.GetString("kafka.group_id"),
			NumWorkers:     viper.GetInt("kafka.num_workers"),
			MinBytes:       viper.GetInt("kafka.min_bytes"),
			MaxBytes:       viper.GetInt("kafka.max_bytes"),
			CommitInterval: viper.GetDuration("kafka.commit_interval"),
			BatchSize:      viper.GetInt("kafka.batch_size"),
			BatchTimeout:   viper.GetDuration("kafka.batch_timeout"),
		},
	}
}
