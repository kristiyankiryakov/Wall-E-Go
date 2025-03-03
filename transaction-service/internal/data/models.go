package data

import "time"

type TransactionType int

const (
	Deposit  TransactionType = 0
	Withdraw TransactionType = 1
)

type Transaction struct {
	ID             int64           `json:"id"`
	WalletID       int64           `json:"wallet_id"`
	Amount         float64         `json:"amount"`
	Type           TransactionType `json:"transaction_type"`
	IdempotencyKey string          `json:"idempotency_key"`
	CreatedAt      time.Time       `json:"created_at"`
}

type DepositRequest struct {
	WalletID       int64   `json:"wallet_id"`
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotency_key"`
}
