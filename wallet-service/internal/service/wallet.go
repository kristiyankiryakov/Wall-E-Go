package service

import (
	"context"
	"strconv"
	walletpb "wallet-service/proto"

	"wallet-service/internal/data"
	"wallet-service/internal/jwt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type WalletService struct {
	walletpb.UnimplementedWalletServiceServer
	walletRepo data.WalletRepository
	jwtUtil    jwt.JWTUtil
}

func NewWalletService(walletRepo data.WalletRepository, jwtUtil jwt.JWTUtil) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
		jwtUtil:    jwtUtil,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, req *walletpb.CreateWalletRequest) (*walletpb.CreateWalletResponse, error) {
	var newWallet data.Wallet
	// Extract token from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["authorization"]) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization header is missing")
	}

	authHeader := md["authorization"][0]
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Validate token and get user_id
	userID, err := s.jwtUtil.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token : %v", err)
	}
	if userID == 0 {
		return nil, status.Errorf(codes.Internal, "error processing token")
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

	return &walletpb.CreateWalletResponse{WalletId: strconv.Itoa(walletID)}, nil
}
