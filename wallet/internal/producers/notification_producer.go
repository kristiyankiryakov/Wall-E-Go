package producers

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
	"wallet/internal/events"
)

type NotificationProducer interface {
	PublishNotificationEvents(ctx context.Context, events []*events.Notification)
}

type NotificationProducerImpl struct {
	writer *kafka.Writer
}

func NewNotificationProducer(addr, topic string, batchSize int, batchTimeout time.Duration) *NotificationProducerImpl {
	return &NotificationProducerImpl{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(addr),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			BatchSize:              batchSize,
			BatchTimeout:           batchTimeout,
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *NotificationProducerImpl) PublishNotificationEvents(ctx context.Context, events []*events.Notification) {
	if len(events) == 0 {
		return
	}
	messages := make([]kafka.Message, 0, len(events))

	for _, e := range events {
		if e == nil {
			log.Printf("Skipping nil notification event")
			continue
		}
		msgBytes, err := json.Marshal(e)
		if err != nil {
			log.Printf("Failed to marshal notification event for transaction %v: %v", e.Data["transaction_id"], err)
			continue
		}
		messages = append(messages, kafka.Message{
			Key:   []byte(e.Data["transaction_id"].(string)),
			Value: msgBytes,
		})
	}

	if len(messages) == 0 {
		log.Printf("No valid notification events to publish")
		return
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		log.Printf("Failed to publish %d notification events: %v", len(messages), err)
	} else {
		log.Printf("Published %d notification events", len(messages))
	}
}

func (p *NotificationProducerImpl) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("Failed to close Kafka writer: %v", err)
	}
}
