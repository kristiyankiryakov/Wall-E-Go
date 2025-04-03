package models

type TransactionRequest struct {
	WalletID       string  `json:"wallet_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	IdempotencyKey string  `json:"idempotency_key" binding:"required"`
}
