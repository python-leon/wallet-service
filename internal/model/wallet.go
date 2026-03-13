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

// TransferRequest 不属于model 一般放在序列化服务里
type TransferRequest struct {
	FromWalletID string `json:"from_wallet_id"`
	ToWalletID   string `json:"to_wallet_id"`
	Amount       int64  `json:"amount"`
}

// TransferResponse represents the result of a transfer
type TransferResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	FromWalletID  string `json:"from_wallet_id,omitempty"`
	ToWalletID    string `json:"to_wallet_id,omitempty"`
	FromBalance   int64  `json:"from_balance,omitempty"`
	ToBalance     int64  `json:"to_balance,omitempty"`
	TransferredAt string `json:"transferred_at,omitempty"`
}
