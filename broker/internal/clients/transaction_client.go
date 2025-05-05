package clients

import (
	"broker/internal/models"
	"broker/proto/gen"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

type TransactionClient struct {
	client gen.TransactionServiceClient
	log    *logrus.Logger
}

func NewTransactionClient(addr string, log *logrus.Logger) (*TransactionClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.WithError(err).Error("Failed to connect to transaction service")
		return nil, fmt.Errorf("failed to connect to transaction service: %w", err)
	}
	return &TransactionClient{
		client: gen.NewTransactionServiceClient(conn),
		log:    log,
	}, nil
}

func (c *TransactionClient) Deposit(ctx context.Context, req models.TransactionRequest) (string, error) {
	c.log.Debug("Depositing money")
	resp, err := c.client.Deposit(ctx, &gen.TransactionRequest{
		WalletId:       req.WalletID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to deposit: %w", err)
	}

	return resp.GetTransactionId(), nil
}
