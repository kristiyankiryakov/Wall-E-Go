package mail

type Notification struct {
	body      string
	recipient string
}

func NewNotification(body, recipient string) *Notification {
	return &Notification{
		body:      body,
		recipient: recipient,
	}
}
func (n *Notification) GetBody() string {
	return n.body
}

func (n *Notification) GetRecipient() string {
	return n.recipient
}
