package service

import (
	"context"
	"log"
	"strconv"
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

	IsWalletOwner(ctx context.Context, req *walletpb.IsOwnerRequest) (*walletpb.IsOwnerResponse, error)
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
	userID, err := s.extractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if wallet with such name already exists for this user
	existingWallet, err := s.walletRepo.GetByUserIdAndWalletName(userID, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error checking for duplicate wallet: %v", err)
	}
	if existingWallet.ID != "" {
		return nil, status.Errorf(codes.AlreadyExists, "wallet with name %v already exists for user with ID: %v", req.Name, userID)
	}

	newWallet.Name = req.Name
	newWallet.UserID = userID

	walletID, err := s.walletRepo.CreateWallet(newWallet)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating a wallet: %v", err)
	}

	return &walletpb.CreateWalletResponse{
		WalletId: walletID,
	}, nil
}

func (s *WalletServiceImpl) ViewBalance(ctx context.Context, req *walletpb.ViewBalanceRequest) (*walletpb.ViewBalanceResponse, error) {
	userID, err := s.extractUserIDFromContext(ctx)
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

func (s *WalletServiceImpl) IsWalletOwner(ctx context.Context, req *walletpb.IsOwnerRequest) (*walletpb.IsOwnerResponse, error) {
	wallet, err := s.getWalletByUserAndWalletID(req.GetUserId(), req.GetWalletId())
	if err != nil {
		return nil, err
	}

	isValid := wallet != nil

	return &walletpb.IsOwnerResponse{
		Valid: isValid,
	}, nil
}

func (s *WalletServiceImpl) getWalletByUserAndWalletID(userID int64, walletID string) (*data.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserIdAndWalletID(userID, walletID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting wallet: %v", err)
	}

	if wallet.ID == "" {
		return nil, status.Errorf(codes.NotFound, "wallet with id: %v does not exists", walletID)
	}

	return wallet, nil
}

func (s *WalletServiceImpl) extractUserIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("err metadata receive")
		return 0, status.Errorf(codes.Internal, "error receiving metadata")
	}
	extracted := md.Get("userID")[0]
	userID, err := strconv.Atoi(extracted)
	if err != nil {
		log.Println(err)
		return 0, status.Errorf(codes.Internal, "error converting userID's type")
	}

	return int64(userID), nil
}
