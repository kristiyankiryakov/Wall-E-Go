package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(topic string) *Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,

		Compression:  kafka.Snappy,
		BatchSize:    1000,                  //MAX
		BatchBytes:   104857,                //1MB
		BatchTimeout: 20 * time.Millisecond, // wait for more messages before sending

		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{writer: writer}
}

func (p *Producer) PublishDepositInitiated(ctx context.Context, walletID string, amount float64, TransactionID string) error {

	event := map[string]interface{}{
		"wallet_id":      walletID,
		"amount":         amount,
		"transaction_id": TransactionID,
	}
	msg, err := json.Marshal(event)
	if err != nil {
		log.Printf("error marshalling message: %s", err)
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(TransactionID),
		Value: msg,
	})
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Println("Failed to close Kafka writer:", err)
	}
}
