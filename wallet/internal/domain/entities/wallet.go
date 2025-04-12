package entities

import (
	"time"
)

type Wallet struct {
	ID        string
	UserID    int64
	Name      string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
