package clients

import (
	"broker/internal/models"
	"broker/proto/gen"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc"
)

type WalletClient struct {
	client gen.WalletServiceClient
	log    *logrus.Logger
}

func NewWalletClient(addr string, log *logrus.Logger) (*WalletClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.WithError(err).Error("Failed to connect to auth service")
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	return &WalletClient{
		client: gen.NewWalletServiceClient(conn),
		log:    log,
	}, nil
}

func (c *WalletClient) CreateWallet(ctx context.Context, walletName string) (string, error) {
	c.log.Debug("Creating wallet")
	resp, err := c.client.CreateWallet(ctx, &gen.CreateWalletRequest{
		Name: walletName,
	})
	if err != nil {
		c.log.WithError(err).Error("Failed to create wallet")
		return "", fmt.Errorf("failed to create wallet: %w", err)
	}

	return resp.WalletId, nil
}

func (c *WalletClient) ViewBalance(ctx context.Context, walletID string) (*models.ViewBalanceResponse, error) {
	c.log.Debug("Viewing balance")
	resp, err := c.client.ViewBalance(ctx, &gen.ViewBalanceRequest{
		WalletId: walletID,
	})
	if err != nil {
		c.log.WithError(err).Error("Failed to view balance")
		return nil, fmt.Errorf("failed to view balance: %w", err)
	}

	return &models.ViewBalanceResponse{
		Name:    resp.GetName(),
		Balance: resp.GetBalance(),
	}, nil
}

func (c *WalletClient) IsWalletOwner(ctx context.Context, userID int64, walletID string) (bool, error) {
	c.log.Debug("Checking wallet ownership")
	resp, err := c.client.IsWalletOwner(ctx, &gen.IsOwnerRequest{
		WalletId: walletID,
		UserId:   int64(userID),
	})
	if err != nil {
		c.log.WithError(err).Error("Failed to check wallet ownership")
		return false, fmt.Errorf("failed to check wallet ownership: %w", err)
	}

	return resp.GetValid(), nil
}

func (c *WalletClient) HealthCheck(ctx context.Context, empty *emptypb.Empty) error {
	c.log.Debug("Checking wallet service health")
	_, err := c.client.HealthCheck(ctx, empty)
	if err != nil {
		c.log.WithError(err).Error("Failed to check wallet service health")
		return fmt.Errorf("failed to check wallet service health: %w", err)
	}
	return nil
}
