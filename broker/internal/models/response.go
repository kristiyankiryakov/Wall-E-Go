package models

type TransactionType string

const (
	Deposit  TransactionType = "DEPOSIT"
	Withdraw TransactionType = "WITHDRAW"
)

type ViewBalanceResponse struct {
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type TransactionResponse struct {
	Message       string             `json:"message,omitempty"`
	TransactionID string             `json:"transaction_id" validate:"required"`
	WalletID      string             `json:"wallet_id" validate:"required"`
	Amount        float64            `json:"amount" validate:"required"`
	Type          TransactionRequest `json:"transaction_type" validate:"required"`
}
