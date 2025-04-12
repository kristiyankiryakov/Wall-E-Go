package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"
	"transaction/internal/domain/entities"
	"transaction/internal/domain/repositories"

	"github.com/segmentio/kafka-go"
)

const TRANSACTION_STATUS_COMPLETED entities.TransactionStatus = "COMPLETED"

type Consumer struct {
	reader          *kafka.Reader
	db              *sql.DB
	transactionRepo *repositories.PostgresTransactionRepository
	batchSize       int
	batchTimeout    time.Duration
}

func NewConsumer(db *sql.DB, topic string) *Consumer {
	transactionRepo := repositories.NewPostgresTransactionRepository(db)
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{"localhost:9092"},
			Topic:          topic,
			GroupID:        "transaction-group",
			MinBytes:       1e3,  // 1KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: 1 * time.Second,
		}),
		db:              db,
		transactionRepo: transactionRepo,
		batchSize:       100,             // Process up to 100 messages in a batch
		batchTimeout:    1 * time.Second, // Process batch every second or when full
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	log.Printf("Starting transaction Kafka consumer with batch size: %d", c.batchSize)

	// Channel to collect transaction IDs
	transactionIDs := make([]string, 0, c.batchSize)

	// Mutex to protect access to the transaction IDs slice
	var mu sync.Mutex

	// Create a ticker for batch processing
	ticker := time.NewTicker(c.batchTimeout)
	defer ticker.Stop()

	// Create a channel to signal batch processing
	processBatch := make(chan struct{})

	// Start a goroutine to process batches
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processBatch <- struct{}{}
			}
		}
	}()

	// Process batches of messages
	for {
		select {
		case <-ctx.Done():
			// Process any remaining transactions before exiting
			c.processBatch(ctx, transactionIDs)
			return

		case <-processBatch:
			// Check if we have any transactions to process
			mu.Lock()
			if len(transactionIDs) > 0 {
				// Process the batch and reset
				ids := make([]string, len(transactionIDs))
				copy(ids, transactionIDs)
				transactionIDs = transactionIDs[:0] // Clear the slice
				mu.Unlock()

				// Process the batch in the background
				go c.processBatch(ctx, ids)
			} else {
				mu.Unlock()
			}

		default:
			// Try to read a message with a short timeout
			readCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			msg, err := c.reader.FetchMessage(readCtx)
			cancel()

			if err != nil {
				if readCtx.Err() != context.DeadlineExceeded {
					log.Printf("Error fetching message: %v", err)
				}
				// No message available, try again after a short sleep
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// Process the message
			var event struct {
				TransactionID string `json:"transaction_id"`
			}

			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Println("Failed to unmarshal event:", err)
				c.reader.CommitMessages(ctx, msg)
				continue
			}

			// Add transaction ID to the batch
			mu.Lock()
			transactionIDs = append(transactionIDs, event.TransactionID)
			currentBatchSize := len(transactionIDs)
			mu.Unlock()

			// Mark message as processed
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}

			// If batch is full, process it immediately
			if currentBatchSize >= c.batchSize {
				processBatch <- struct{}{}
			}
		}
	}
}

// processBatch updates the status of a batch of transactions
func (c *Consumer) processBatch(ctx context.Context, transactionIDs []string) {
	if len(transactionIDs) == 0 {
		return
	}

	log.Printf("Processing batch of %d transaction status updates", len(transactionIDs))
	start := time.Now()

	// Use the concurrent update method from the repository
	err := c.transactionRepo.UpdateStatusConcurrently(ctx, transactionIDs, TRANSACTION_STATUS_COMPLETED)
	if err != nil {
		log.Printf("Error updating transaction statuses: %v", err)
	} else {
		elapsed := time.Since(start)
		log.Printf("Successfully updated %d transactions in %v", len(transactionIDs), elapsed)
	}
}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Println("Failed to close Kafka reader:", err)
	}
}
