package clients

import (
	"broker-service/internal/models"
	txpb "broker-service/proto"
	"context"
	"log"

	"google.golang.org/grpc"
)

type TransactionClient struct {
	client txpb.TransactionServiceClient
}

func NewTransactionClient(addr string) (*TransactionClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.Printf("Failed to connect to auth service: %v", err)
		return nil, err
	}
	return &TransactionClient{client: txpb.NewTransactionServiceClient(conn)}, nil
}

func (c *TransactionClient) Deposit(ctx context.Context, req models.DepositRequest) (int64, error) {
	resp, err := c.client.Deposit(ctx, &txpb.DepositRequest{
		WalletId:       req.WalletID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return 0, err
	}

	return resp.GetTransactionId(), nil
}
