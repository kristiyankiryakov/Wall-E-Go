package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer() *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("kafka:9092"),
		Topic:    "deposit-initiated",
		Balancer: &kafka.LeastBytes{},

		BatchTimeout: 50 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(context.Background(), kafka.Message{Value: []byte("test")})
		if err == nil {
			log.Println("Successfully connected to Kafka")
			break
		}
		log.Printf("Kafka not ready, retrying (%d/10): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	return &Producer{writer: writer}
}

func (p *Producer) PublishDepositInitiated(ctx context.Context, walletID int64, amount float64, TransactionID int64) error {
	event := map[string]interface{}{
		"wallet_id":      walletID,
		"amount":         amount,
		"transaction_id": TransactionID,
	}
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(TransactionID, 10)),
		Value: msg,
	})
}

func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Println("Failed to close Kafka writer:", err)
	}
}
