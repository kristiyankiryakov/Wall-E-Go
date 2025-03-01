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

func (c *WalletClient) CreateWallet(ctx context.Context, walletName string) (string, error) {
	resp, err := c.client.CreateWallet(ctx, &walletpb.CreateWalletRequest{
		Name: walletName,
	})
	if err != nil {
		return "", err
	}
	return resp.WalletId, nil
}

func (c *WalletClient) ViewBalance(ctx context.Context, walletID string) (*models.ViewBalanceResponse, error) {
	rpcResponse, err := c.client.ViewBalance(ctx, &walletpb.ViewBalanceRequest{
		WalletId: walletID,
	})
	if err != nil {
		return nil, err
	}

	log.Println(rpcResponse)

	return &models.ViewBalanceResponse{
		Name:    rpcResponse.GetName(),
		Balance: rpcResponse.GetBalance(),
	}, nil
}
