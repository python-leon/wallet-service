package model

import (
	"time"
)

// Wallet represents a user's wallet with balance
type Wallet struct {
	ID        string    `json:"id"`
	Balance   int64     `json:"balance"` // 余额，以最小单位（分）存储，避免浮点数误差
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
