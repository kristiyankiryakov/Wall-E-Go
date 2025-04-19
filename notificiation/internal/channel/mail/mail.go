package mail

import (
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"notification/internal/channel"
	"notification/internal/config"
)

type Mail struct {
	Config *config.Mail
}

func (m *Mail) Send(n channel.Notification) error {
	newMetadata := n.GetMetadata()
	template, ok := newMetadata["template"]
	if !ok {
		return errors.New("template not found in metadata")
	}
	switch template {
	case "deposit":
		if err := m.DepositTemplate(n.GetMetadata()); err != nil {
			return errors.New("failed to send deposit template: " + err.Error())
		}
	default:
		return errors.New("template not found")
	}

	return nil
}

func NewMail(cfg *config.Mail) *Mail {
	return &Mail{
		Config: cfg,
	}
}

func (m *Mail) DepositTemplate(meta map[string]any) error {
	e := email.NewEmail()
	e.From = "wall-e-go@gmail.com"
	e.To = []string{"recipient@tobeadded.com"}
	e.Subject = "Deposit Notification"

	amount, ok := meta["amount"].(float64)
	if !ok {
		return errors.New("amount not found in metadata")
	}
	walletID, ok := meta["wallet_id"].(string)
	if !ok {
		return errors.New("wallet_id not found in metadata")
	}
	transactionID, ok := meta["transaction_id"].(string)
	if !ok {
		return errors.New("transaction_id not found in metadata")
	}

	e.Text = []byte(fmt.Sprintf("Deposit of %.2f, with transactionID: %s to wallet %s was successful, you", amount, transactionID, walletID))

	return e.Send(m.Config.SMTPHost+":"+m.Config.SMTPPort, m.Config.Auth)
}
