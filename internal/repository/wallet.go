package repository

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/python-leon/wallet-service/internal/model"
)

// WalletRepository defines the interface for wallet storage operations
type WalletRepository interface {
	Create() *model.Wallet
	GetByID(id string) (*model.Wallet, bool)
	UpdateBalance(id string, newBalance int64) bool
	Transfer(fromID, toID string, amount int64) (*TransferResult, error)
}

// TransferResult represents the result of a transfer operation
type TransferResult struct {
	FromWallet *model.Wallet
	ToWallet   *model.Wallet
	Success    bool
	Message    string
}

// InMemoryWalletRepository implements WalletRepository using in-memory storage
type InMemoryWalletRepository struct {
	wallets map[string]*model.Wallet
	mu      sync.RWMutex
}

// NewInMemoryWalletRepository creates a new in-memory wallet repository
func NewInMemoryWalletRepository() *InMemoryWalletRepository {
	return &InMemoryWalletRepository{
		wallets: make(map[string]*model.Wallet),
	}
}

// Create creates a new wallet with zero balance
func (r *InMemoryWalletRepository) Create() *model.Wallet {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	// 分布式系统下可更换使用雪花算法生成ID
	wallet := &model.Wallet{
		ID:        uuid.New().String(),
		Balance:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.wallets[wallet.ID] = wallet
	return wallet
}

func (r *InMemoryWalletRepository) GetByID(id string) (*model.Wallet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wallet, exists := r.wallets[id]
	if !exists {
		return nil, false
	}
	// Return a copy to prevent external modifications
	walletCopy := *wallet
	return &walletCopy, true
}

// UpdateBalance updates the balance of a wallet
func (r *InMemoryWalletRepository) UpdateBalance(id string, newBalance int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	wallet, exists := r.wallets[id]
	if !exists {
		return false
	}

	wallet.Balance = newBalance
	wallet.UpdatedAt = time.Now()
	return true
}

// Transfer transfers amount from one wallet to another atomically
func (r *InMemoryWalletRepository) Transfer(fromID, toID string, amount int64) (*TransferResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fromWallet, fromExists := r.wallets[fromID]
	toWallet, toExists := r.wallets[toID]

	if !fromExists {
		return &TransferResult{
			Success: false,
			Message: "source wallet not found",
		}, nil
	}

	if !toExists {
		return &TransferResult{
			Success: false,
			Message: "destination wallet not found",
		}, nil
	}

	if amount <= 0 {
		return &TransferResult{
			Success: false,
			Message: "amount must be positive",
		}, nil
	}

	if fromWallet.Balance < amount {
		return &TransferResult{
			Success: false,
			Message: "insufficient balance",
		}, nil
	}

	// Perform the transfer
	now := time.Now()
	fromWallet.Balance -= amount
	fromWallet.UpdatedAt = now
	toWallet.Balance += amount
	toWallet.UpdatedAt = now

	// Return copies
	fromCopy := *fromWallet
	toCopy := *toWallet

	return &TransferResult{
		FromWallet: &fromCopy,
		ToWallet:   &toCopy,
		Success:    true,
		Message:    "transfer successful",
	}, nil
}
