package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"
	"wallet/internal/producer"

	"github.com/segmentio/kafka-go"
)

type DepositEvent struct {
	WalletID      string  `json:"wallet_id"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id"`
}

type Consumer struct {
	reader           *kafka.Reader
	db               *sql.DB
	batchSize        int
	maxWorkers       int
	processingPeriod time.Duration
}

// ConsumerConfig allows customization of consumer behavior
type ConsumerConfig struct {
	BatchSize        int
	MaxWorkers       int
	ProcessingPeriod time.Duration
}

// DefaultConsumerConfig provides sensible defaults
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		BatchSize:        50,                     // Process up to 50 messages at once
		MaxWorkers:       10,                     // Up to 10 concurrent workers
		ProcessingPeriod: 500 * time.Millisecond, // Check for messages every 500ms
	}
}

func NewConsumer(db *sql.DB, readerTopic string) *Consumer {
	return NewConsumerWithConfig(db, readerTopic, DefaultConsumerConfig())
}

func NewConsumerWithConfig(db *sql.DB, readerTopic string, config ConsumerConfig) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{"localhost:9092"},
			Topic:          readerTopic,
			GroupID:        "wallet-group",
			MinBytes:       1e3,  // 1KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: 1 * time.Second,
		}),
		db:               db,
		batchSize:        config.BatchSize,
		maxWorkers:       config.MaxWorkers,
		processingPeriod: config.ProcessingPeriod,
	}
}

func (c *Consumer) Consume(ctx context.Context, producer *producer.Producer) {
	log.Printf("Starting Kafka consumer with max workers: %d, batch size: %d", c.maxWorkers, c.batchSize)

	// Channel for passing messages to workers
	taskChan := make(chan kafka.Message, c.batchSize*2)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go c.startWorker(ctx, i, taskChan, &wg, producer)
	}

	// Periodically fetch batches of messages
	ticker := time.NewTicker(c.processingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(taskChan) // Signal workers to stop
			wg.Wait()       // Wait for all workers to finish
			return
		case <-ticker.C:
			c.fetchBatch(ctx, taskChan)
		}
	}
}

// fetchBatch reads a batch of messages from Kafka and sends them to the worker pool
func (c *Consumer) fetchBatch(ctx context.Context, taskChan chan<- kafka.Message) {
	// Create a context with timeout for batch reading
	batchCtx, cancel := context.WithTimeout(ctx, c.processingPeriod)
	defer cancel()

	// Read batch of messages
	batch := make([]kafka.Message, 0, c.batchSize)
	for i := 0; i < c.batchSize; i++ {
		msg, err := c.reader.FetchMessage(batchCtx)
		if err != nil {
			// Context deadline exceeded means no more messages available
			if batchCtx.Err() != nil {
				break
			}
			log.Printf("Error fetching message: %v", err)
			break
		}
		batch = append(batch, msg)
	}

	// Send messages to workers
	if len(batch) > 0 {
		log.Printf("Fetched batch of %d messages", len(batch))
		for _, msg := range batch {
			select {
			case <-ctx.Done():
				return
			case taskChan <- msg:
				// Message sent to worker
			}
		}
	}
}

// startWorker processes messages from the task channel
func (c *Consumer) startWorker(ctx context.Context, id int, taskChan <-chan kafka.Message, wg *sync.WaitGroup, producer *producer.Producer) {
	defer wg.Done()
	log.Printf("Starting worker %d", id)

	// Map to collect completion events for batch publishing
	completionEvents := make(map[string]string)

	// Ticker for periodic batch commits
	commitTicker := time.NewTicker(200 * time.Millisecond)
	defer commitTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Commit any remaining events before exiting
			producer.PublishCompletionEvents(ctx, completionEvents)
			return

		case msg, ok := <-taskChan:
			if !ok {
				// Channel closed, exit worker
				producer.PublishCompletionEvents(ctx, completionEvents)
				return
			}

			// Process the message
			transactionID, ok := c.processMessage(ctx, msg)
			if ok {
				completionEvents[transactionID] = transactionID

				// Commit message to mark as processed
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("Worker %d: Failed to commit message: %v", id, err)
				}

				// If we've accumulated enough events, publish them as a batch
				if len(completionEvents) >= 10 {
					producer.PublishCompletionEvents(ctx, completionEvents)
					completionEvents = make(map[string]string)
				}
			}

		case <-commitTicker.C:
			// Periodically publish any accumulated events
			if len(completionEvents) > 0 {
				producer.PublishCompletionEvents(ctx, completionEvents)
				completionEvents = make(map[string]string)
			}
		}
	}
}

// processMessage handles a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) (string, bool) {
	var event DepositEvent

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return "", false
	}

	// Start a transaction
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return "", false
	}

	// Update balance
	_, err = tx.ExecContext(ctx,
		"UPDATE wallets SET balance = balance + $1 WHERE id = $2",
		event.Amount, event.WalletID)
	if err != nil {
		log.Printf("Failed to update balance: %v", err)
		tx.Rollback()
		return "", false
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return "", false
	}

	return event.TransactionID, true
}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Printf("Failed to close Kafka reader: %v", err)
	}
}
