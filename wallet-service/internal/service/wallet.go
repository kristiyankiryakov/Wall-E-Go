package service

import (
	"context"
	walletpb "wallet-service/proto"

	"wallet-service/internal/data"
	"wallet-service/internal/jwt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type WalletService interface {
	CreateWallet(ctx context.Context, req *walletpb.CreateWalletRequest) (*walletpb.CreateWalletResponse, error)
	ViewBalance(ctx context.Context, req *walletpb.ViewBalanceRequest) (*walletpb.ViewBalanceResponse, error)
}

type WalletServiceImpl struct {
	walletpb.UnimplementedWalletServiceServer
	walletRepo data.WalletRepository
	jwtUtil    jwt.JWTUtil
}

func NewWalletService(walletRepo data.WalletRepository, jwtUtil jwt.JWTUtil) *WalletServiceImpl {
	return &WalletServiceImpl{
		walletRepo: walletRepo,
		jwtUtil:    jwtUtil,
	}
}

func (s *WalletServiceImpl) CreateWallet(ctx context.Context, req *walletpb.CreateWalletRequest) (*walletpb.CreateWalletResponse, error) {
	var newWallet data.Wallet

	userID, err := s.extractUserIdAndValidateToken(ctx)
	if err != nil {
		return nil, err
	}

	// Check if wallet with such name already exists for this user
	existingWallet, err := s.walletRepo.GetByUserIdAndWalletName(userID, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error checking for duplicate wallet: %v", err)
	}
	if existingWallet.ID != 0 {
		return nil, status.Errorf(codes.AlreadyExists, "wallet with name %v already exists for user with ID: %v", req.Name, userID)
	}

	newWallet.Name = req.Name
	newWallet.UserID = userID

	walletID, err := s.walletRepo.CreateWallet(newWallet)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating a wallet: %v", err)
	}

	return &walletpb.CreateWalletResponse{
		WalletId: int64(walletID),
	}, nil
}

func (s *WalletServiceImpl) ViewBalance(ctx context.Context, req *walletpb.ViewBalanceRequest) (*walletpb.ViewBalanceResponse, error) {
	userID, err := s.extractUserIdAndValidateToken(ctx)
	if err != nil {
		return nil, err
	}

	wallet, err := s.getWalletByUserAndWalletID(userID, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &walletpb.ViewBalanceResponse{
		Balance: wallet.Balance,
		Name:    wallet.Name,
	}, nil

}

func (s *WalletServiceImpl) getWalletByUserAndWalletID(userID, walletID int64) (*data.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserIdAndWalletID(int64(userID), walletID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting wallet: %v", err)
	}
	if wallet.ID == 0 {
		return nil, status.Errorf(codes.NotFound, "wallet with id: %d does not exists", walletID)
	}

	return wallet, nil
}

func (s *WalletServiceImpl) extractUserIdAndValidateToken(ctx context.Context) (int64, error) {

	// Extract token from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["authorization"]) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "authorization header is missing")
	}

	authHeader := md["authorization"][0]
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Validate token and get user_id
	userID, err := s.jwtUtil.ValidateToken(token)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token : %v", err)
	}
	if userID == 0 {
		return 0, status.Errorf(codes.Internal, "error processing token")
	}

	return int64(userID), nil
}
