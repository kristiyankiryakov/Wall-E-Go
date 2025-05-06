package handlers

import (
	"broker/internal/clients"
	"broker/internal/utils"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WalletHandler interface {
	CreateWallet(c *gin.Context)
	ViewBalance(c *gin.Context)
	HealthCheck(c *gin.Context)
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

	ctx := c.Request.Context()

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "error handling walletID"})
		return
	}

	ctx := c.Request.Context()

	response, err := h.walletClient.ViewBalance(ctx, walletID)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *WalletHandlerImpl) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.walletClient.HealthCheck(ctx, &emptypb.Empty{})
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
