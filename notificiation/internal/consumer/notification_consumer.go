package consumer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"notification/internal/channel/mail"

	"notification/internal/config"
	"notification/internal/service"
	"notification/logger"
	"sync"
	"time"
)

type NotificationEvent struct {
	Channel string
	Data    map[string]any
}

var log = logger.NewLogger()

type Consumer struct {
	reader       *kafka.Reader
	sender       service.NotificationService
	batchSize    int
	batchTimeout time.Duration
	numWorkers   int
}

func NewConsumer(cfg *config.Kafka, sender service.NotificationService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			Topic:          cfg.Topic,
			GroupID:        cfg.GroupID,
			MinBytes:       cfg.MinBytes,
			MaxBytes:       cfg.MaxBytes,
			CommitInterval: cfg.CommitInterval,
		}),
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
		sender:       sender,
		numWorkers:   cfg.NumWorkers,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	wg := &sync.WaitGroup{}

	// Create a channel for messages to be processed
	notifications := make(chan kafka.Message, c.batchSize)

	// Start workers
	for i := 0; i < c.numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerID)

			for msg := range notifications {

				// Unmarshal the message to NotificationEvent
				var event NotificationEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					continue
				}

				mailNotification := mail.NewNotification(event.Data)
				if err := c.sender.SendNotification(ctx, mailNotification); err != nil {
					log.Printf("Failed to send notification for message: %v", err)
				} else {
					log.Printf("Successfully processed message: offset=%d", msg.Offset)
				}

				// Commit the message
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("Failed to commit message: %v", err)
				}
			}

			log.Printf("Worker %d stopped", workerID)
		}(i)
	}

	// Read messages from Kafka and send to workers
	go func() {
		defer close(notifications)

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping consumer")
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					// If context was cancelled, exit gracefully
					if ctx.Err() != nil {
						return
					}
					// Otherwise wait a bit and continue
					time.Sleep(time.Second)
					continue
				}

				notifications <- msg
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Println("Failed to close Kafka reader:", err)
	}
}
