package handlers

import (
	"broker-service/internal/clients"
	"broker-service/internal/utils"
	"net/http"
	"strconv"

	"github.com/kristiyankiryakov/Wall-E-Go-Common/dto"

	"github.com/gin-gonic/gin"
)

type TransactionHandler interface {
	Deposit(*gin.Context)
}

type TransactionHandlerImpl struct {
	transactionClient *clients.TransactionClient
}

func NewTransactionHandler(transactionClient *clients.TransactionClient) *TransactionHandlerImpl {
	return &TransactionHandlerImpl{
		transactionClient: transactionClient,
	}
}

func (h *TransactionHandlerImpl) Deposit(c *gin.Context) {
	var req dto.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	walletID, err := strconv.Atoi(c.Query("walletID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error handling walletID"})
		return
	}
	convertedWalletID := int64(walletID)
	req.WalletID = &convertedWalletID

	ctx := c.Request.Context()

	txID, err := h.transactionClient.Deposit(ctx, req)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction_id": txID})
}
