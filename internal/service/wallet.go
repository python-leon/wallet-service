package service

import (
	"github.com/python-leon/wallet-service/internal/errors"
	"github.com/python-leon/wallet-service/internal/model"
	"github.com/python-leon/wallet-service/internal/repository"
)

type WalletService interface {
	CreateWallet() *model.Wallet
	GetWallet(id string) (*model.Wallet, error)
}

// walletService implements WalletService
type walletService struct {
	repo repository.WalletRepository
}

// NewWalletService creates a new wallet service
func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{
		repo: repo,
	}
}

// CreateWallet creates a new wallet with zero balance
func (s *walletService) CreateWallet() *model.Wallet {
	return s.repo.Create()
}

// GetWallet retrieves a wallet by ID
func (s *walletService) GetWallet(id string) (*model.Wallet, error) {
	wallet, exists := s.repo.GetByID(id)
	if !exists {
		return nil, errors.ErrWalletNotFound
	}
	return wallet, nil
}
