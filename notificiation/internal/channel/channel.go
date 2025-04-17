package channel

type NotificationType string

const (
	Email NotificationType = "email"
	SMS   NotificationType = "sms"
	Push  NotificationType = "push"
)

type Notification interface {
	GetBody() string
	GetRecipient() string // Add this method
	GetType() NotificationType
	GetMetadata() map[string]string
}

type MessageSender interface {
	Send(n Notification) error // Update signature to match implementation
}
