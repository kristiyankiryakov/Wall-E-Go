package auth

import (
	"auth/internal/jwt"
	"auth/internal/user"
	"auth/proto/gen"
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	RegisterUser(ctx context.Context, req *gen.RegisterUserRequest) (*gen.RegisterUserResponse, error)
	Authenticate(ctx context.Context, req *gen.AuthenticateRequest) (*gen.AuthenticateResponse, error)
}

type AuthServiceImpl struct {
	gen.UnimplementedAuthServiceServer
	jwtUtil  jwt.JWTUtil
	userRepo user.UserRepository
	log      *logrus.Logger
}

func NewAuthService(jwtUtil jwt.JWTUtil, userRepo user.UserRepository, log *logrus.Logger) *AuthServiceImpl {
	return &AuthServiceImpl{
		jwtUtil:  jwtUtil,
		userRepo: userRepo,
		log:      log,
	}
}

func (s *AuthServiceImpl) RegisterUser(ctx context.Context, req *gen.RegisterUserRequest) (*gen.RegisterUserResponse, error) {
	user := user.User{Username: req.Username, Password: req.Password}

	// Check for existing user
	if err := s.handleExistingUser(ctx, user.Username); err != nil {
		s.log.WithError(err).Warn("attempt to register existing user")
		return nil, err // Return gRPC error directly
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		s.log.WithError(err).Error("failed to hash password")
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	token, err := s.jwtUtil.GenerateToken(user.ID)
	if err != nil {
		s.log.WithError(err).Error("failed to generate token")
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	if _, err := s.userRepo.Insert(ctx, user); err != nil {
		s.log.WithError(err).Error("failed to insert user into database")
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &gen.RegisterUserResponse{Token: token}, nil
}

func (s *AuthServiceImpl) handleExistingUser(ctx context.Context, username string) error {
	existingUser, err := s.userRepo.GetByUsername(ctx, username)

	if err == nil && existingUser.ID != 0 {
		return status.Errorf(codes.AlreadyExists, "user %s already exists", username)
	} else if err != nil {
		s.log.WithError(err).Error("failed to check user existence")
		return status.Errorf(codes.Internal, "failed to check user existence: %v", err)
	}

	return nil
}

func (s *AuthServiceImpl) Authenticate(ctx context.Context, req *gen.AuthenticateRequest) (*gen.AuthenticateResponse, error) {
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.log.WithError(err).Error("failed to find user")
		return nil, status.Errorf(codes.Internal, "failed find user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		s.log.WithError(err).Warn("password mismatch")
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.jwtUtil.GenerateToken(existingUser.ID)
	if err != nil {
		s.log.WithError(err).Error("failed to generate token")
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	s.log.WithField("username", existingUser.Username).Info("user authenticated successfully")
	return &gen.AuthenticateResponse{
		Token: token,
	}, nil
}
