package channel

type Notification interface {
	GetBody() string
	GetRecipient() string
}

type Channel interface {
	SendMessage(n Notification) error
}
