package services

import (
	"database/sql"
	"wall-e-go/auth-service/internal/models"
	"wall-e-go/auth-service/internal/repository"
	jwt "wall-e-go/auth-service/internal/util"
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

	if err := authService.handleExistingUser(newUser.Username); err != nil {
		return "", err
	}

	if err := authService.createUser(&newUser); err != nil {
		return "", err
	}

	token, err := authService.generateToken(newUser.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (authService AuthService) handleExistingUser(username string) error {
	existingUser, err := authService.UserRepository.GetUserByUsername(username)
	if err != nil && err != sql.ErrNoRows {
		return errors.WrapError(errors.ErrInternal, "Error checking user existence")
	}
	if existingUser != nil {
		return errors.ErrAlreadyExists
	}
	return nil
}

func (authService AuthService) createUser(newUser *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WrapError(errors.ErrInternal, "Error hashing password")
	}
	newUser.Password = string(hashedPassword)

	if err := authService.UserRepository.CreateUser(*newUser); err != nil {
		return errors.WrapError(errors.ErrInternal, "Error creating user")
	}
	return nil
}

func (authService AuthService) generateToken(username string) (string, error) {
	jwtUtil := jwt.JWTUtil{}
	token, err := jwtUtil.GenerateToken(username)
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error generating token")
	}

	return token, nil
}
