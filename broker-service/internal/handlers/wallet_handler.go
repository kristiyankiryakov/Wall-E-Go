package handlers

import (
	"broker-service/internal/clients"
	"broker-service/internal/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type WalletHandler interface {
}

type WalletHandlerImpl struct {
	walletClient *clients.WalletClient
}

func NewWalletHandler(walletClient *clients.WalletClient) *WalletHandlerImpl {
	return &WalletHandlerImpl{
		walletClient: walletClient,
	}
}

func (h *WalletHandlerImpl) CreateWallet(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	// Extract JWT from HTTP Header
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	ctx := metadata.AppendToOutgoingContext(c.Request.Context(), "authorization", authHeader)

	walletID, err := h.walletClient.CreateWallet(ctx, req.Name)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet_id": walletID})
}

func (h *WalletHandlerImpl) ViewBalance(c *gin.Context) {
	walletID := c.Query("walletID")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid walletID: %v", walletID)})
		return
	}

	// Extract JWT from HTTP Header
	authHeader := c.Request.Header.Get("Authorization")
	ctx := metadata.AppendToOutgoingContext(c.Request.Context(), "authorization", authHeader)

	response, err := h.walletClient.ViewBalance(ctx, walletID)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
