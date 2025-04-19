package events

type Notification struct {
	Channel string
	Data    map[string]any
}

type Deposit struct {
	WalletID      string  `json:"wallet_id"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id"`
}
