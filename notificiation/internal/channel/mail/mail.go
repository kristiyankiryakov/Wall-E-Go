package mail

import (
	"github.com/jordan-wright/email"
	"notification/internal/channel"
	"notification/internal/config"
)

type Mail struct {
	Config *config.Mail
}

func (m *Mail) Send(n channel.Notification) error {
	e := email.NewEmail()
	e.From = "Kris Tst <kris-test@example.com>"
	e.To = []string{n.GetRecipient()}
	e.Subject = "Test Subject"
	e.Text = []byte(n.GetBody())

	return e.Send(m.Config.SMTPHost+":"+m.Config.SMTPPort, m.Config.Auth)
}

func NewMail(cfg *config.Mail) *Mail {
	return &Mail{
		Config: cfg,
	}
}
