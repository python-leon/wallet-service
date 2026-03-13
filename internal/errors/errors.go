package errors

import "errors"

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrSameWallet          = errors.New("source and destination wallets must be different")
)
