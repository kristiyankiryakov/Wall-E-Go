package transaction

import "time"

type TransactionType string

const (
	Deposit  TransactionType = "DEPOSIT"
	Withdraw TransactionType = "WITHDRAW"
)

type TransactionStatus string

type Transaction struct {
	ID             string          `db:"id"`
	WalletID       string          `db:"wallet_id"`
	Amount         float64         `db:"amount"`
	Type           TransactionType `db:"transaction_type"`
	IdempotencyKey string          `db:"idempotency_key"`
	Status         string          `db:"status"`
	updatedAt      time.Time       `db:"updated_at"`
	CreatedAt      time.Time       `db:"created_at"`
}
