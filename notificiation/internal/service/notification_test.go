package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"notification/internal/channel"
	"testing"
)

// mockNotification mocks channel.Notification
type mockNotification struct {
	notificationType channel.NotificationType
	metadata         map[string]any
}

func (m *mockNotification) GetType() channel.NotificationType {
	return m.notificationType
}
func (m *mockNotification) GetMetadata() map[string]any {
	return m.metadata
}

// mockSender mocks channel.MessageSender
type mockSender struct {
	sendCalled       bool
	sentNotification channel.Notification
	sendError        error
}

func (m *mockSender) Send(notification channel.Notification) error {
	m.sendCalled = true
	m.sentNotification = notification
	return m.sendError
}

func TestNotificationService_AddChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(ns NotificationService)
		channelType   string
		sender        channel.MessageSender
		expectedError error
	}{
		{
			name:          "when adding a new channel, it must succeed with no error",
			setup:         func(ns NotificationService) {},
			channelType:   "email",
			sender:        &mockSender{},
			expectedError: nil,
		},
		{
			name: "when adding an existing channel, it must return an error",
			setup: func(ns NotificationService) {
				ns.AddChannel("email", &mockSender{})
			},
			channelType:   "email",
			sender:        &mockSender{},
			expectedError: errors.New("channel already registered: email"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNotificationService()
			tt.setup(ns)

			err := ns.AddChannel(tt.channelType, tt.sender)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				// Verify the channel was added
				sender, err := ns.GetChannel(tt.channelType)
				assert.NoError(t, err)
				assert.Equal(t, tt.sender, sender)
			}
		})
	}
}

func TestNotificationService_SendNotification(t *testing.T) {
	t.Parallel()

	// Helper to create a cancelled context
	cancelledContext := func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}

	tests := []struct {
		name             string
		channelType      string
		sender           *mockSender
		addChannel       bool
		ctx              context.Context
		notification     channel.Notification
		expectedError    error
		expectSendCalled bool
	}{
		{
			name:             "when sending through a registered channel, it must succeed",
			channelType:      "email",
			sender:           &mockSender{},
			addChannel:       true,
			ctx:              context.Background(),
			notification:     &mockNotification{notificationType: "email"},
			expectedError:    nil,
			expectSendCalled: true,
		},
		{
			name:             "when sending through an unregistered channel, it must return an error",
			channelType:      "sms",
			sender:           nil,
			addChannel:       false,
			ctx:              context.Background(),
			notification:     &mockNotification{notificationType: "sms"},
			expectedError:    errors.New("channel not registered: sms"),
			expectSendCalled: false,
		},
		{
			name:             "when sending with a cancelled context, it must return context error",
			channelType:      "email",
			sender:           &mockSender{},
			addChannel:       true,
			ctx:              cancelledContext(),
			notification:     &mockNotification{notificationType: "email"},
			expectedError:    context.Canceled,
			expectSendCalled: false,
		},
		{
			name:             "when sender fails, it must propagate the error",
			channelType:      "email",
			sender:           &mockSender{sendError: errors.New("send failed")},
			addChannel:       true,
			ctx:              context.Background(),
			notification:     &mockNotification{notificationType: "email"},
			expectedError:    errors.New("send failed"),
			expectSendCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNotificationService()
			if tt.addChannel {
				ns.AddChannel(tt.channelType, tt.sender)
			}

			err := ns.SendNotification(tt.ctx, tt.notification)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.sender != nil {
				assert.Equal(t, tt.expectSendCalled, tt.sender.sendCalled)
				if tt.expectSendCalled {
					assert.Equal(t, tt.notification, tt.sender.sentNotification)
				}
			}
		})
	}
}
