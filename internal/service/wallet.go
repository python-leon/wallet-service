package service

import (
	"time"

	"github.com/python-leon/wallet-service/internal/errors"
	"github.com/python-leon/wallet-service/internal/model"
	"github.com/python-leon/wallet-service/internal/repository"
)

type WalletService interface {
	CreateWallet() *model.Wallet
	GetWallet(id string) (*model.Wallet, error)
	Deposit(id string, amount int64) (*model.Wallet, error)
	Transfer(req *model.TransferRequest) (*model.TransferResponse, error)
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

// Deposit adds funds to a wallet
func (s *walletService) Deposit(id string, amount int64) (*model.Wallet, error) {
	if amount <= 0 {
		return nil, errors.ErrInvalidAmount
	}

	wallet, exists := s.repo.GetByID(id)
	if !exists {
		return nil, errors.ErrWalletNotFound
	}

	newBalance := wallet.Balance + amount
	if !s.repo.UpdateBalance(id, newBalance) {
		return nil, errors.ErrInsufficientBalance
	}

	// 获取更新后的钱包信息
	updatedWallet, exists := s.repo.GetByID(id)
	if !exists {
		return nil, errors.ErrWalletNotFound
	}

	return updatedWallet, nil
}

// Transfer transfers funds from one wallet to another
func (s *walletService) Transfer(req *model.TransferRequest) (*model.TransferResponse, error) {
	// Validate request
	if req.FromWalletID == "" {
		return &model.TransferResponse{
			Success: false,
			Message: "source wallet ID is required",
		}, nil
	}

	if req.ToWalletID == "" {
		return &model.TransferResponse{
			Success: false,
			Message: "destination wallet ID is required",
		}, nil
	}

	if req.FromWalletID == req.ToWalletID {
		return &model.TransferResponse{
			Success: false,
			Message: "cannot transfer to the same wallet",
		}, nil
	}

	if req.Amount <= 0 {
		return &model.TransferResponse{
			Success: false,
			Message: "amount must be positive",
		}, nil
	}

	result, err := s.repo.Transfer(req.FromWalletID, req.ToWalletID, req.Amount)
	if err != nil {
		return nil, err
	}

	response := &model.TransferResponse{
		Success: result.Success,
		Message: result.Message,
	}

	if result.Success {
		response.FromWalletID = result.FromWallet.ID
		response.ToWalletID = result.ToWallet.ID
		response.FromBalance = result.FromWallet.Balance
		response.ToBalance = result.ToWallet.Balance
		response.TransferredAt = time.Now().Format(time.RFC3339)
	}

	return response, nil
}
