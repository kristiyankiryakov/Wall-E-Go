package models

type ViewBalanceRequest struct {
	WalletID int64 `json:"wallet_id"`
}

type DepositRequest struct {
	WalletID       int64   `json:"wallet_id"`
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotency_key"`
}
