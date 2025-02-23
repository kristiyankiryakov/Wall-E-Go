package data

import "time"

type TransactionType int

const (
	Deposit  TransactionType = 1
	Withdraw TransactionType = 2
)

type Wallet struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletTransaction struct {
	ID        int             `json:"id"`
	Amount    float64         `json:"amount"`
	Type      TransactionType `json:"transaction_type"`
	WalletID  int             `json:"wallet_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
