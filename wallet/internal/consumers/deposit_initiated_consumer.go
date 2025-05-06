package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
	"wallet/internal/events"
	"wallet/internal/producers"

	"github.com/segmentio/kafka-go"
)

const (
	NOTIFICATION_CHANNEL = "email"
	EMAIL_TEMPLATE       = "deposit"
)

type Config struct {
	Brokers        []string
	Topic          string
	GroupID        string
	BatchSize      int
	MinBytes       int
	MaxBytes       int
	CommitInterval time.Duration
}

type Consumer struct {
	reader          *kafka.Reader
	db              *pgxpool.Pool
	depositProducer producers.DepositCompletedProducer
	notifyProducer  producers.NotificationProducer
	batchSize       int
}

func NewConsumer(db *pgxpool.Pool, cfg *Config, depositProducer producers.DepositCompletedProducer, notifyProducer producers.NotificationProducer) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			Topic:          cfg.Topic,
			GroupID:        cfg.GroupID,
			MinBytes:       cfg.MinBytes,
			MaxBytes:       cfg.MaxBytes,
			CommitInterval: cfg.CommitInterval,
		}),
		db:              db,
		depositProducer: depositProducer,
		notifyProducer:  notifyProducer,
		batchSize:       cfg.BatchSize,
	}
}

// Consume processes Kafka messages and handles deposits.
func (c *Consumer) Consume(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic: %s with batch size: %d", c.reader.Config().Topic, c.batchSize)

	for {
		// Fetch a batch of messages
		messages, err := c.fetchBatch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("Consumer stopped: %v", ctx.Err())
				return
			}
			log.Printf("Error fetching batch: %v", err)
			continue
		}
		if len(messages) == 0 {
			continue
		}

		// Process the batch
		depositEvents, notificationEvents := c.processBatch(ctx, messages)
		if len(depositEvents) == 0 {
			log.Printf("No valid events in batch of %d messages", len(messages))
			continue
		}

		// Publish to deposit-completed and notification topics
		c.depositProducer.PublishDepositCompletedEvents(ctx, depositEvents)
		c.notifyProducer.PublishNotificationEvents(ctx, notificationEvents)

		// Commit the batch
		if err := c.reader.CommitMessages(ctx, messages...); err != nil {
			log.Printf("Failed to commit batch of %d messages: %v", len(messages), err)
		} else {
			log.Printf("Committed batch of %d messages", len(messages))
		}
	}
}

func (c *Consumer) fetchBatch(ctx context.Context) ([]kafka.Message, error) {
	messages := make([]kafka.Message, 0, c.batchSize)

	// Try to fetch up to batchSize messages with a reasonable timeout
	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	for len(messages) < c.batchSize {
		msg, err := c.reader.FetchMessage(fetchCtx)
		if err != nil {
			// If timeout and we have some messages, return them
			if errors.Is(err, context.DeadlineExceeded) && len(messages) > 0 {
				return messages, nil
			}

			// If no messages and timeout, return empty with nil error to avoid error logging
			if errors.Is(err, context.DeadlineExceeded) {
				time.Sleep(500 * time.Millisecond) // Add a small delay to prevent CPU spinning
				return nil, nil
			}

			// For other errors, return the error
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// processBatch handles a batch of Kafka messages in a single database transaction.
func (c *Consumer) processBatch(ctx context.Context, messages []kafka.Message) ([]*events.Deposit, []*events.Notification) {
	if len(messages) == 0 {
		return nil, nil
	}

	// Start a database transaction
	tx, err := c.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, nil
	}

	depositEvents := make([]*events.Deposit, 0, len(messages))
	notificationEvents := make([]*events.Notification, 0, len(messages))

	for _, msg := range messages {
		var event events.Deposit
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		// Update wallet balance
		_, err = tx.Exec(ctx, "UPDATE wallets SET balance = balance + $1 WHERE id = $2", event.Amount, event.WalletID)
		if err != nil {
			log.Printf("Failed to update balance for transaction %s: %v", event.TransactionID, err)
			continue
		}

		// Collect successful events
		depositEvents = append(depositEvents, &event)
		notificationEvents = append(notificationEvents, &events.Notification{
			Channel: NOTIFICATION_CHANNEL,
			Data: map[string]any{
				"wallet_id":      event.WalletID,
				"amount":         event.Amount,
				"transaction_id": event.TransactionID,
				"template":       EMAIL_TEMPLATE,
			},
		})
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		tx.Rollback(ctx)
		return nil, nil
	}

	return depositEvents, notificationEvents
}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Printf("Failed to close Kafka reader: %v", err)
	}
}
