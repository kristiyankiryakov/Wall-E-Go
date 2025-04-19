package service

import (
	"context"
	"errors"
	"notification/internal/channel"
	"notification/logger"
)

var log = logger.NewLogger()

type NotificationService interface {
	// SendNotification sends a notification through the appropriate channel
	SendNotification(ctx context.Context, notification channel.Notification) error

	// AddChannel registers a new notification channel
	AddChannel(channelType string, sender channel.MessageSender) error

	// GetChannel retrieves a registered notification channel
	GetChannel(channelType string) (channel.MessageSender, error)
}

func NewNotificationService() NotificationService {
	return &notificationService{
		channels: make(map[string]channel.MessageSender),
	}
}

type notificationService struct {
	channels map[string]channel.MessageSender
}

func (n *notificationService) SendNotification(ctx context.Context, notification channel.Notification) error {
	// Get channel type from notification
	channelType := string(notification.GetType())

	// Check if channel exists
	sender, exists := n.channels[channelType]
	if !exists {
		return errors.New("channel not registered: " + channelType)
	}

	// Send notification with context awareness
	log.Printf("Sending notification via %s channel", channelType)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := sender.Send(notification); err != nil {
			log.Printf("Failed to send notification: %v", err)
			return err
		}
		return nil
	}
}

func (n *notificationService) AddChannel(channelType string, sender channel.MessageSender) error {
	if _, exists := n.channels[channelType]; exists {
		return errors.New("channel already registered: " + channelType)
	}

	n.channels[channelType] = sender
	return nil
}

func (n *notificationService) GetChannel(channelType string) (channel.MessageSender, error) {
	sender, exists := n.channels[channelType]
	if !exists {
		return nil, errors.New("channel not registered: " + channelType)
	}

	return sender, nil
}
