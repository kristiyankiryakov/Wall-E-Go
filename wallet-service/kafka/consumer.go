package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

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
			log.Println("Failed to read Kafka message:", err)
			continue
		}

		var event struct {
			WalletID      int64   `json:"wallet_id"`
			Amount        float64 `json:"amount"`
			TransactionID int64   `json:"transaction_id"`
		}

		log.Println(string(msg.Value))

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("Failed to unmarshal event:", err)
			continue
		}

		// Update balance
		_, err = c.db.ExecContext(ctx,
			"UPDATE wallets SET balance = balance + $1 WHERE id = $2",
			event.Amount, event.WalletID)
		if err != nil {
			log.Println("Failed to update balance:", err)
			continue
		}

		// Publish completion event
		completionEvent := map[string]int64{"transaction_id": event.TransactionID}
		msgBytes, _ := json.Marshal(completionEvent)
		c.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(strconv.FormatInt(event.TransactionID, 10)),
			Value: msgBytes,
		})
	}
}
