package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/python-leon/wallet-service/internal/model"
	"gorm.io/gorm"
)

// DBWalletRepository implements WalletRepository using PostgreSQL
type DBWalletRepository struct {
	db *gorm.DB
}

// NewDBWalletRepository creates a new database-backed wallet repository
func NewDBWalletRepository(db *gorm.DB) *DBWalletRepository {
	return &DBWalletRepository{db: db}
}

// AutoMigrate runs auto migration for the wallet model
func (r *DBWalletRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&WalletModel{})
}

// WalletModel is the GORM model for wallet table
type WalletModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	Balance   int64     `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for WalletModel
func (WalletModel) TableName() string {
	return "wallets"
}

// Create creates a new wallet with zero balance
func (r *DBWalletRepository) Create() *model.Wallet {
	walletModel := &WalletModel{
		ID:      uuid.New().String(),
		Balance: 0,
	}

	if err := r.db.Create(walletModel).Error; err != nil {
		return nil
	}

	return &model.Wallet{
		ID:        walletModel.ID,
		Balance:   walletModel.Balance,
		CreatedAt: walletModel.CreatedAt,
		UpdatedAt: walletModel.UpdatedAt,
	}
}

// GetByID retrieves a wallet by ID
func (r *DBWalletRepository) GetByID(id string) (*model.Wallet, bool) {
	var walletModel WalletModel
	if err := r.db.Where("id = ?", id).First(&walletModel).Error; err != nil {
		return nil, false
	}

	return &model.Wallet{
		ID:        walletModel.ID,
		Balance:   walletModel.Balance,
		CreatedAt: walletModel.CreatedAt,
		UpdatedAt: walletModel.UpdatedAt,
	}, true
}

// UpdateBalance updates the balance of a wallet
func (r *DBWalletRepository) UpdateBalance(id string, newBalance int64) bool {
	result := r.db.Model(&WalletModel{}).Where("id = ?", id).Update("balance", newBalance)
	return result.RowsAffected > 0
}

// Deposit adds funds to a wallet atomically
func (r *DBWalletRepository) Deposit(id string, amount int64) (*model.Wallet, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	var walletModel WalletModel

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get wallet with lock
		if err := tx.Where("id = ?", id).First(&walletModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("wallet not found")
			}
			return err
		}

		// Calculate new balance
		newBalance := walletModel.Balance + amount
		now := time.Now()

		// Update database
		if err := tx.Model(&WalletModel{}).Where("id = ?", id).Updates(map[string]interface{}{
			"balance":    newBalance,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		// Update local model to reflect the new state
		walletModel.Balance = newBalance
		walletModel.UpdatedAt = now

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &model.Wallet{
		ID:        walletModel.ID,
		Balance:   walletModel.Balance,
		CreatedAt: walletModel.CreatedAt,
		UpdatedAt: walletModel.UpdatedAt,
	}, nil
}

// Transfer transfers amount from one wallet to another atomically
func (r *DBWalletRepository) Transfer(fromID, toID string, amount int64) (*TransferResult, error) {
	if amount <= 0 {
		return &TransferResult{
			Success: false,
			Message: "amount must be positive",
		}, nil
	}

	var fromWallet, toWallet WalletModel

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get source wallet with lock (ORDER BY id to prevent deadlock)
		if err := tx.Where("id = ?", fromID).First(&fromWallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("source wallet not found")
			}
			return err
		}

		// Get destination wallet with lock
		if err := tx.Where("id = ?", toID).First(&toWallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("destination wallet not found")
			}
			return err
		}

		// Check balance
		if fromWallet.Balance < amount {
			return fmt.Errorf("insufficient balance")
		}

		// Perform transfer
		fromWallet.Balance -= amount
		toWallet.Balance += amount

		if err := tx.Save(&fromWallet).Error; err != nil {
			return err
		}
		if err := tx.Save(&toWallet).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "source wallet not found":
			return &TransferResult{
				Success: false,
				Message: "source wallet not found",
			}, nil
		case "destination wallet not found":
			return &TransferResult{
				Success: false,
				Message: "destination wallet not found",
			}, nil
		case "insufficient balance":
			return &TransferResult{
				Success: false,
				Message: "insufficient balance",
			}, nil
		default:
			return nil, err
		}
	}

	return &TransferResult{
		FromWallet: &model.Wallet{
			ID:        fromWallet.ID,
			Balance:   fromWallet.Balance,
			CreatedAt: fromWallet.CreatedAt,
			UpdatedAt: fromWallet.UpdatedAt,
		},
		ToWallet: &model.Wallet{
			ID:        toWallet.ID,
			Balance:   toWallet.Balance,
			CreatedAt: toWallet.CreatedAt,
			UpdatedAt: toWallet.UpdatedAt,
		},
		Success: true,
		Message: "transfer successful",
	}, nil
}
