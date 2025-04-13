package producer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP("localhost:9092"),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			BatchSize:              100,
			BatchTimeout:           20 * time.Millisecond,
			AllowAutoTopicCreation: true,
		},
	}
}

// PublishCompletionEvents publishes a batch of completion events to Kafka
func (p *Producer) PublishCompletionEvents(ctx context.Context, events map[string]string) {
	if len(events) == 0 {
		return
	}

	messages := make([]kafka.Message, 0, len(events))
	for txID := range events {
		completionEvent := map[string]string{
			"transaction_id": txID,
		}
		msgBytes, _ := json.Marshal(completionEvent)
		messages = append(messages, kafka.Message{
			Key:   []byte(txID),
			Value: msgBytes,
		})
	}

	// Write all messages in a batch
	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		log.Printf("Failed to publish completion events: %v", err)
	} else {
		log.Printf("Published %d completion events", len(messages))
	}
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("Failed to close Kafka writer: %v", err)
	}
}
