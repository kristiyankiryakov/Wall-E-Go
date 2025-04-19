package channel

type NotificationType string

const (
	Email NotificationType = "email"
	SMS   NotificationType = "sms"
	Push  NotificationType = "push"
)

type Notification interface {
	GetType() NotificationType
	GetMetadata() map[string]any
}

type MessageSender interface {
	Send(n Notification) error
}
