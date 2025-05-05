package user

import (
	"time"
)

// User is the structure which holds one user from the database.
type User struct {
	ID        int
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
