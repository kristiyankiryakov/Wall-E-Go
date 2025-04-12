package handlers

import (
	"broker-service/internal/clients"
	"broker-service/internal/models"
	"broker-service/internal/utils"
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
	reqValue, exists := c.Get("transactionRequest")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction request not found"})
		return
	}

	req, ok := reqValue.(models.TransactionRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid transaction request type"})
		return
	}

	ctx := c.Request.Context()

	txID, err := h.transactionClient.Deposit(ctx, req)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction_id": txID, "status": "transaction initiated successfully"})
}
