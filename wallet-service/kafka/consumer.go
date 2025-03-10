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
	writer *kafka.Writer
}

func NewConsumer(db *sql.DB, readerTopic string, writerTopic string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"kafka:9092"},
			Topic:   readerTopic,
			GroupID: "wallet-group",
		}),
		db: db,
		writer: &kafka.Writer{
			Addr:     kafka.TCP("kafka:9092"),
			Topic:    writerTopic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Failed to read Kafka message: %v", err)
			continue
		}

		var event struct {
			WalletID      string  `json:"wallet_id"`
			Amount        float64 `json:"amount"`
			TransactionID string  `json:"transaction_id"`
		}

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		// Start a transaction
		tx, err := c.db.BeginTx(ctx, nil)
		if err != nil {
			log.Println("Failed to begin transaction:", err)
			continue
		}

		// Update balance
		_, err = tx.ExecContext(ctx,
			"UPDATE wallets SET balance = balance + $1 WHERE id = $2",
			event.Amount, event.WalletID)
		if err != nil {
			log.Printf("Failed to update balance: %v", err)
			tx.Rollback()
			continue
		}

		// Commit transaction before Kafka
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction: %v", err)
			continue
		}

		// Publish completion event
		completionEvent := map[string]string{"transaction_id": event.TransactionID}
		msgBytes, _ := json.Marshal(completionEvent)
		c.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(event.TransactionID),
			Value: msgBytes,
		})
	}
}
