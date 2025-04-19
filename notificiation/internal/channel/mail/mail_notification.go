package mail

import "notification/internal/channel"

type Notification struct {
	metadata map[string]any
}

func NewNotification(metadata map[string]any) channel.Notification {
	return &Notification{
		metadata: metadata,
	}
}

func (n *Notification) GetType() channel.NotificationType {
	return channel.Email
}

func (n *Notification) GetMetadata() map[string]any {
	return n.metadata
}
