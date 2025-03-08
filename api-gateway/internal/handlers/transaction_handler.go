package handlers

import (
	"broker-service/internal/clients"
	"broker-service/internal/models"
	"broker-service/internal/utils"
	"fmt"
	"net/http"

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
	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request : %v", err)})
		return
	}
	req.WalletID = c.Query("walletID")

	ctx := c.Request.Context()

	txID, err := h.transactionClient.Deposit(ctx, req)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction_id": txID, "status": "transaction initiated successfully"})
}
