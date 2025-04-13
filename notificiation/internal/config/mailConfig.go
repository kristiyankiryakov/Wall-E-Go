package config

import "net/smtp"

type MailConfig struct {
	SMTPHost string
	SMTPPort string
	Auth     smtp.Auth
}

func NewMailConfig(smtpHost, smtpPort string, auth smtp.Auth) *MailConfig {
	return &MailConfig{
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
		Auth:     auth,
	}
}
