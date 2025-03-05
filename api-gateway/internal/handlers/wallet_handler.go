package handlers

import (
	"broker-service/internal/clients"
	"broker-service/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WalletHandler interface {
	CreateWallet(c *gin.Context)
	ViewBalance(c *gin.Context)
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
	walletID, err := strconv.Atoi(c.Query("walletID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error handling walletID"})
		return
	}

	ctx := c.Request.Context()

	response, err := h.walletClient.ViewBalance(ctx, int64(walletID))
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
