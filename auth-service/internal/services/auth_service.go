package services

import (
	"database/sql"
	"wall-e-go/auth-service/internal/models"
	"wall-e-go/auth-service/internal/repository"
	errors "wall-e-go/common"

	"golang.org/x/crypto/bcrypt"
)

type JWTUtil interface {
	GenerateToken(username string) (string, error)
}

type AuthService struct {
	UserRepository repository.UserRepository
	jwtUtil        JWTUtil
}

func NewAuthService(userRepo repository.UserRepository, jwtUtil JWTUtil) *AuthService {
	return &AuthService{
		UserRepository: userRepo,
		jwtUtil:        jwtUtil,
	}
}

func (as AuthService) RegisterUser(newUser models.User) (string, error) {
	if err := as.handleExistingUser(newUser.Username); err != nil {
		return "", err
	}

	if err := as.createUser(&newUser); err != nil {
		return "", err
	}

	token, err := as.generateToken(newUser.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (as AuthService) handleExistingUser(username string) error {
	existingUser, err := as.UserRepository.GetUserByUsername(username)

	if err != nil && err != sql.ErrNoRows {
		return errors.WrapError(errors.ErrInternal, "Error checking user existence")
	}

	if existingUser != nil {
		return errors.ErrAlreadyExists
	}

	return nil
}

func (as AuthService) createUser(newUser *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WrapError(errors.ErrInternal, "Error hashing password")
	}
	newUser.Password = string(hashedPassword)

	if err := as.UserRepository.CreateUser(*newUser); err != nil {
		return errors.WrapError(errors.ErrInternal, "Error creating user")
	}

	return nil
}

func (as AuthService) generateToken(username string) (string, error) {

	token, err := as.jwtUtil.GenerateToken(username)
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error generating token")
	}

	return token, nil
}

//TODO: add login functionality -> check credentials if good generate a token
// For Auth required services check how it's done - whether a common middleware is shared or each service implement's it's own
