package handlers

import (
	"net/http"
	"wall-e-go/auth-service/internal/models"
	"wall-e-go/auth-service/internal/services"
	errors "wall-e-go/common"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.Error(errors.WrapError(errors.ErrBadRequest, "Invalid input data"))
		return
	}

	token, err := h.AuthService.RegisterUser(user)
	if err != nil {
		// Pass the error to the middleware
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "token": token})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.Error(errors.WrapError(errors.ErrBadRequest, "Invalid input data"))
		return
	}

	token, err := h.AuthService.Login(user)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successful authentication", "token": token})
}
