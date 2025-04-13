package config

import "net/smtp"

type MailConfig struct {
	SMTPHost string
	SMTPPort string
	Auth     smtp.Auth
}
