package clients

import (
	"broker-service/proto/gen"
	"context"
	"log"

	"google.golang.org/grpc"
)

type AuthClient struct {
	client gen.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.Printf("Failed to connect to auth service: %v", err)
		return nil, err
	}
	return &AuthClient{client: gen.NewAuthServiceClient(conn)}, nil
}

func (c *AuthClient) RegisterUser(username, password string) (string, error) {
	resp, err := c.client.RegisterUser(context.Background(), &gen.RegisterUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

func (c *AuthClient) Authenticate(username, password string) (string, error) {
	resp, err := c.client.Authenticate(context.Background(), &gen.AuthenticateRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}
