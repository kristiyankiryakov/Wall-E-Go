package mail

import "notification/internal/channel"

type Notification struct {
	body      string
	recipient string
}

func NewNotification(body, recipient string) *Notification {
	return &Notification{
		body:      body,
		recipient: recipient, // This was missing
	}
}

func (n *Notification) GetBody() string {
	return n.body
}

func (n *Notification) GetRecipient() string {
	return n.recipient
}

func (n *Notification) GetType() channel.NotificationType {
	return channel.Email
}

func (n *Notification) GetMetadata() map[string]string {
	return map[string]string{}
}
