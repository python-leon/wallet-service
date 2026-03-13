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
