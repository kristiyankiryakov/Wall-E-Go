package service

import (
	"context"
	"log"
	"wall-e-go/internal/data"
	"wall-e-go/internal/jwt"
	authpb "wall-e-go/proto"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	RegisterUser(ctx context.Context, req *authpb.RegisterUserRequest) (*authpb.RegisterUserResponse, error)
	Authenticate(ctx context.Context, req *authpb.AuthenticateRequest) (*authpb.AuthenticateResponse, error)
}

type AuthServiceImpl struct {
	authpb.UnimplementedAuthServiceServer
	jwtUtil  jwt.JWTUtil
	userRepo data.UserRepository
}

func NewAuthService(jwtUtil jwt.JWTUtil, userRepo data.UserRepository) *AuthServiceImpl {
	return &AuthServiceImpl{
		jwtUtil:  jwtUtil,
		userRepo: userRepo,
	}
}

func (s *AuthServiceImpl) RegisterUser(ctx context.Context, req *authpb.RegisterUserRequest) (*authpb.RegisterUserResponse, error) {
	user := data.User{Username: req.Username, Password: req.Password}

	// Check for existing user
	if err := s.handleExistingUser(user.Username); err != nil {
		return nil, err // Return gRPC error directly
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	log.Printf("Hashed password: %s", hashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	token, err := s.jwtUtil.GenerateToken(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	if _, err := s.userRepo.Insert(user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &authpb.RegisterUserResponse{Token: token}, nil
}

func (s *AuthServiceImpl) handleExistingUser(username string) error {
	existingUser, err := s.userRepo.GetByUsername(username)

	if err == nil && existingUser.ID != 0 {
		return status.Errorf(codes.AlreadyExists, "user %s already exists", username)
	} else if err != nil {
		return status.Errorf(codes.Internal, "failed to check user existence: %v", err)
	}

	return nil
}

func (s *AuthServiceImpl) Authenticate(ctx context.Context, req *authpb.AuthenticateRequest) (*authpb.AuthenticateResponse, error) {
	existingUser, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed find user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.jwtUtil.GenerateToken(existingUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &authpb.AuthenticateResponse{Token: token}, nil
}
