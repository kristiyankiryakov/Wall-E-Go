package clients

import (
	"broker/internal/models"
	"broker/proto/gen"
	"context"
	"log"

	"google.golang.org/grpc"
)

type TransactionClient struct {
	client gen.TransactionServiceClient
}

func NewTransactionClient(addr string) (*TransactionClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.Printf("Failed to connect to auth service: %v", err)
		return nil, err
	}
	return &TransactionClient{client: gen.NewTransactionServiceClient(conn)}, nil
}

func (c *TransactionClient) Deposit(ctx context.Context, req models.TransactionRequest) (string, error) {
	resp, err := c.client.Deposit(ctx, &gen.TransactionRequest{
		WalletId:       req.WalletID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return "", err
	}

	return resp.GetTransactionId(), nil
}
