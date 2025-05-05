package clients

import (
	"broker/proto/gen"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AuthClient struct {
	client gen.AuthServiceClient
	log    *logrus.Logger
}

func NewAuthClient(addr string, log *logrus.Logger) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
	if err != nil {
		log.WithError(err).Error("Failed to connect to auth service")
		return nil, err
	}
	return &AuthClient{
		client: gen.NewAuthServiceClient(conn),
		log:    log,
	}, nil
}

func (c *AuthClient) RegisterUser(username, password string) (string, error) {
	c.log.WithFields(logrus.Fields{
		"username": username,
	}).Debug("Registering user")

	resp, err := c.client.RegisterUser(context.Background(), &gen.RegisterUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		c.log.WithError(err).Error("Failed to register user")
		return "", fmt.Errorf("failed to register user: %w", err)
	}
	return resp.Token, nil
}

func (c *AuthClient) Authenticate(username, password string) (string, error) {
	c.log.WithFields(logrus.Fields{
		"username": username,
	}).Debug("Authenticating user")

	resp, err := c.client.Authenticate(context.Background(), &gen.AuthenticateRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		c.log.WithError(err).Error("Failed to authenticate user")
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}
	return resp.Token, nil
}
