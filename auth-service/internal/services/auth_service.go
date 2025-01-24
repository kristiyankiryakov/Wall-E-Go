package services

import (
	"wall-e-go/auth-service/internal/models"
	"wall-e-go/auth-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{UserRepository: userRepo}
}

func (authService AuthService) RegisterUser(newUser models.User) error {

	//TODO: resource already exists exception handling

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser.Password = string(hashedPassword)

	return authService.UserRepository.CreateUser(newUser)
}
