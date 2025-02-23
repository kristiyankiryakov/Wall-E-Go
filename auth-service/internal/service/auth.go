package service

import (
	"context"
	"log"
	"wall-e-go/internal/data"
	"wall-e-go/internal/jwt"
	pb "wall-e-go/proto"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	jwtUtil  jwt.JWTUtil
	userRepo data.UserRepository
}

func NewAuthService(jwtUtil jwt.JWTUtil, userRepo data.UserRepository) *AuthService {
	return &AuthService{
		jwtUtil:  jwtUtil,
		userRepo: userRepo,
	}
}

type JWTUtil interface {
	GenerateToken(username string) (string, error)
}

func (s *AuthService) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
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

	token, err := s.jwtUtil.GenerateToken(user.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	if _, err := s.userRepo.Insert(user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.RegisterUserResponse{Token: token}, nil
}

func (s *AuthService) handleExistingUser(username string) error {
	existingUser, err := s.userRepo.GetByUsername(username)

	if err == nil && existingUser.ID != 0 {
		return status.Errorf(codes.AlreadyExists, "user %s already exists", username)
	} else if err != nil {
		return status.Errorf(codes.Internal, "failed to check user existence: %v", err)
	}

	return nil
}

func (s *AuthService) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	existingUser, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed find user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.jwtUtil.GenerateToken(existingUser.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &pb.AuthenticateResponse{Token: token}, nil
}
