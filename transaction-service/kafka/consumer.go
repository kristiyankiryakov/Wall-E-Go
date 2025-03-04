package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	db     *sql.DB
}

func NewConsumer(db *sql.DB) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"kafka:9092"},
			Topic:   "deposit-completed",
			GroupID: "transaction-group",
		}),
		db: db,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Println("Failed to read Kafka message:", err)
			continue
		}

		var event struct {
			TransactionID int64 `json:"transaction_id"`
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("Failed to unmarshal event:", err)
			continue
		}

		_, err = c.db.ExecContext(ctx,
			`UPDATE transactions SET status = 'COMPLETED' WHERE id = $1 AND status = 'PENDING'`,
			event.TransactionID)
		if err != nil {
			log.Println("Failed to update transaction status:", err)
		}
	}
}

func (c *Consumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Println("Failed to close Kafka reader:", err)
	}
}
