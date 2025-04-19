package producers

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
	"wallet/internal/events"
)

type DepositCompletedProducer interface {
	PublishDepositCompletedEvents(ctx context.Context, events []*events.Deposit)
}

type DepositCompletedProducerImpl struct {
	writer *kafka.Writer
}

func NewDepositCompletedProducer(addr, topic string, batchSize int, batchTimeout time.Duration) *DepositCompletedProducerImpl {
	return &DepositCompletedProducerImpl{
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

func (p *DepositCompletedProducerImpl) PublishDepositCompletedEvents(ctx context.Context, events []*events.Deposit) {
	if len(events) == 0 {
		return
	}

	messages := make([]kafka.Message, 0, len(events))
	for _, e := range events {
		if e == nil {
			log.Printf("Skipping nil deposit event")
			continue
		}
		msgBytes, err := json.Marshal(e)
		if err != nil {
			log.Printf("Failed to marshal deposit event for transaction %s: %v", e.TransactionID, err)
			continue
		}
		messages = append(messages, kafka.Message{
			Key:   []byte(e.TransactionID),
			Value: msgBytes,
		})
	}

	if len(messages) == 0 {
		log.Printf("No valid deposit events to publish")
		return
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		log.Printf("Failed to publish %d deposit-completed events: %v", len(messages), err)
	} else {
		log.Printf("Published %d deposit-completed events", len(messages))
	}
}

func (p *DepositCompletedProducerImpl) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("Failed to close Kafka writer: %v", err)
	}
}
