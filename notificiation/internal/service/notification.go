package service

import "notification/internal/channel"

type NotificationService interface {
	SendNotification(notification channel.Notification) error
}
