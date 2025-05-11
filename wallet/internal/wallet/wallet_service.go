package wallet

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"strconv"
	"wallet/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	CreateWallet(ctx context.Context, req *gen.CreateWalletRequest) (*gen.CreateWalletResponse, error)
	ViewBalance(ctx context.Context, req *gen.ViewBalanceRequest) (*gen.ViewBalanceResponse, error)
	HealthCheck(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error)
}

type service struct {
	gen.UnimplementedWalletServiceServer
	repo Repository
	log  *logrus.Logger
}

func NewWalletService(repo Repository, log *logrus.Logger) *service {
	return &service{
		repo: repo,
		log:  log,
	}
}

func (s *service) HealthCheck(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("Health check called")
	return req, nil
}

func (s *service) CreateWallet(ctx context.Context, req *gen.CreateWalletRequest) (*gen.CreateWalletResponse, error) {
	var newWallet Wallet

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.log.Error("no metadata provided in context")
		return nil, status.Error(codes.Unauthenticated, "no metadata provided")
	}

	userIDStr := md.Get("userID")
	if len(userIDStr) == 0 {
		s.log.Error("user ID not found in metadata")
		return nil, status.Error(codes.Unauthenticated, "user ID not found in metadata")
	}

	userID, err := strconv.Atoi(userIDStr[0])
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID")
	}

	// Check if wallet with such name already exists for this user
	existingWallet, err := s.repo.GetByUserIdAndWalletName(ctx, userID, req.Name)
	if err != nil {
		s.log.Errorf("error checking for duplicate wallet: %v", err)
		return nil, status.Errorf(codes.Internal, "error handling create wallet request")
	}
	if existingWallet.ID != "" {
		return nil, status.Errorf(codes.AlreadyExists, "wallet with name %v already exists for user with ID: %v", req.Name, userID)
	}

	newWallet.Name = req.Name
	newWallet.UserID = userID

	walletID, err := s.repo.CreateWallet(ctx, &newWallet)
	if err != nil {
		s.log.Errorf("error creating wallet: %v", err)
		return nil, status.Errorf(codes.Internal, "error creating a wallet")
	}

	return &gen.CreateWalletResponse{
		WalletId: walletID,
	}, nil
}

func (s *service) ViewBalance(ctx context.Context, req *gen.ViewBalanceRequest) (*gen.ViewBalanceResponse, error) {
	log.Println("CreateWallet called with request: ", req, " and context: ", ctx)
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}

	wallet, err := s.repo.GetByUserIdAndWalletID(ctx, userID, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &gen.ViewBalanceResponse{
		Balance: wallet.Balance,
		Name:    wallet.Name,
	}, nil
}
