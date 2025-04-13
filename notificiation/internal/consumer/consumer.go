package consumer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type Consumer struct {
	reader       *kafka.Reader
	batchSize    int
	batchTimeout time.Duration
}

func NewConsumer(topic, group string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{"localhost:9092"},
			Topic:          topic,
			GroupID:        group,
			MinBytes:       1e3,  // 1KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: 1 * time.Second,
		}),
		batchSize:    100,             // Process up to 100 messages in a batch
		batchTimeout: 1 * time.Second, // Process batch every second or when full
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	//wg := &sync.WaitGroup{}
	//const numWorkers = 5

}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Println("Failed to close Kafka reader:", err)
	}
}
