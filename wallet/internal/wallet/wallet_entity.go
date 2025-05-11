package wallet

import (
	"time"
)

type Wallet struct {
	ID        string
	UserID    int
	Name      string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
