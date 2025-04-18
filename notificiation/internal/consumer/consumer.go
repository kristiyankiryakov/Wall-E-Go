package consumer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"notification/internal/channel/mail"
	"notification/internal/config"
	"notification/internal/service"
	"notification/logger"
	"sync"
	"time"
)

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
			MinBytes:       1e3,  // 1KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: 1 * time.Second,
		}),
		batchSize:    100,             // Process up to 100 messages in a batch
		batchTimeout: 1 * time.Second, // Process batch every second or when full
		sender:       sender,
		numWorkers:   cfg.NumWorkers,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	wg := &sync.WaitGroup{}

	// Create a channel for messages to be processed
	messageCh := make(chan kafka.Message, c.batchSize)

	// Start workers
	for i := 0; i < c.numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerID)

			for msg := range messageCh {
				// Process message and send email notification
				//TODO: fix hardcoded value
				n := mail.NewNotification(
					"Test body",
					"emineo@abv.bg",
				)
				if err := c.sender.SendNotification(ctx, n); err != nil {
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
		defer close(messageCh)

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

				messageCh <- msg
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
