package entities

import "time"

type TransactionType int

type TransactionStatus string

type Transaction struct {
	ID             string
	WalletID       string
	Amount         float64
	Type           TransactionType
	IdempotencyKey string
	updatedAt      time.Time
	CreatedAt      time.Time
}
