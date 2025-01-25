package services

import (
	"database/sql"
	"wall-e-go/auth-service/internal/models"
	"wall-e-go/auth-service/internal/repository"
	jwt_util "wall-e-go/auth-service/internal/util"
	errors "wall-e-go/common"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{UserRepository: userRepo}
}

func (authService AuthService) RegisterUser(newUser models.User) (string, error) {

	existingUser, err := authService.UserRepository.GetUserByUsername(newUser.Username)
	if err != nil && err != sql.ErrNoRows {
		return "", errors.WrapError(errors.ErrInternal, "Error checking user existence")
	}
	if existingUser != nil {
		return "", errors.ErrAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error hashing password")
	}
	newUser.Password = string(hashedPassword)

	if err := authService.UserRepository.CreateUser(newUser); err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error creating user")
	}

	jwtUtil := jwt_util.JWTUtil{}
	token, err := jwtUtil.GenerateToken(newUser.Username)
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error generating token")
	}

	return token, nil
}
