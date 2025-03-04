package handlers

import (
	"broker-service/internal/clients"
	"broker-service/internal/models"
	"broker-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
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
	var req models.DepositRequest

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

	txID, err := h.transactionClient.Deposit(ctx, req)
	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction_id": txID})
}
