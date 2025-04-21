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

type notification struct {
	metadata map[string]any
}

func NewNotification(metadata map[string]any) Notification {
	return &notification{
		metadata: metadata,
	}
}

func (n *notification) GetType() NotificationType {
	return Email
}

func (n *notification) GetMetadata() map[string]any {
	return n.metadata
}

type MessageSender interface {
	Send(n Notification) error
}
