package clients

import (
	"broker-service/internal/models"
	walletpb "broker-service/proto"
	"context"
	"log"

	"google.golang.org/grpc"
)

type WalletClient struct {
	client walletpb.WalletServiceClient
}

func NewWalletClient(addr string) (*WalletClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.Printf("Failed to connect to auth service: %v", err)
		return nil, err
	}
	return &WalletClient{client: walletpb.NewWalletServiceClient(conn)}, nil
}

func (c *WalletClient) CreateWallet(ctx context.Context, walletName string) (int64, error) {
	resp, err := c.client.CreateWallet(ctx, &walletpb.CreateWalletRequest{
		Name: walletName,
	})
	if err != nil {
		return 0, err
	}

	return resp.WalletId, nil
}

func (c *WalletClient) ViewBalance(ctx context.Context, walletID int64) (*models.ViewBalanceResponse, error) {
	resp, err := c.client.ViewBalance(ctx, &walletpb.ViewBalanceRequest{
		WalletId: walletID,
	})
	if err != nil {
		return nil, err
	}

	return &models.ViewBalanceResponse{
		Name:    resp.GetName(),
		Balance: resp.GetBalance(),
	}, nil
}
