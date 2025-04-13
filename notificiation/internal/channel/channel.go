package channel

type Notification interface {
	GetBody() string
	GetRecipient() string
}

type MessageSender interface {
	Send(n Notification) error
}
