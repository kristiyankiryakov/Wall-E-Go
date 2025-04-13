package handlers

import (
	"broker/internal/clients"
	"broker/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Authenticate(c *gin.Context)
}

type AuthHandlerImpl struct {
	authClient *clients.AuthClient
}

func NewAuthHandler(authClient *clients.AuthClient) *AuthHandlerImpl {
	return &AuthHandlerImpl{
		authClient: authClient,
	}
}

func (h *AuthHandlerImpl) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.authClient.RegisterUser(req.Username, req.Password)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandlerImpl) Authenticate(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.authClient.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
