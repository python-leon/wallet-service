package service

import (
	"testing"

	"github.com/python-leon/wallet-service/internal/model"
	"github.com/python-leon/wallet-service/internal/repository"
)

// MockWalletRepository 是一个模拟的仓库层实现，用于测试
type MockWalletRepository struct {
	CreateFunc func() *model.Wallet
}

func (m *MockWalletRepository) GetByID(id string) (*model.Wallet, bool) {
	//TODO implement me
	panic("implement me")
}

func (m *MockWalletRepository) Create() *model.Wallet {
	if m.CreateFunc != nil {
		return m.CreateFunc()
	}
	// 默认返回一个示例钱包
	return &model.Wallet{
		ID:      "mock-wallet-id",
		Balance: 0,
	}
}

// 测试 CreateWallet 方法
func TestWalletService_CreateWallet(t *testing.T) {
	// 创建模拟仓库
	mockRepo := &MockWalletRepository{
		CreateFunc: func() *model.Wallet {
			return &model.Wallet{
				ID:      "new-wallet-id-from-repo",
				Balance: 0,
			}
		},
	}

	// 创建服务实例
	svc := NewWalletService(mockRepo)

	// 调用创建钱包方法
	wallet := svc.CreateWallet()

	// 验证返回的钱包
	if wallet == nil {
		t.Error("期望返回一个钱包，但得到 nil")
	}

	if wallet.ID != "new-wallet-id-from-repo" {
		t.Errorf("期望钱包ID为 'new-wallet-id-from-repo'，但得到 '%v'", wallet.ID)
	}

	if wallet.Balance != 0 {
		t.Errorf("期望钱包余额为 0，但得到 %v", wallet.Balance)
	}
}

// 测试使用真实仓库的集成
func TestWalletService_CreateWallet_Integration(t *testing.T) {
	// 使用真实的仓库
	repo := repository.NewInMemoryWalletRepository()
	svc := NewWalletService(repo)

	// 调用创建钱包方法
	wallet := svc.CreateWallet()

	// 验证返回的钱包
	if wallet == nil {
		t.Error("期望返回一个钱包，但得到 nil")
	}

	if wallet.ID == "" {
		t.Error("期望钱包有ID，但ID为空")
	}

	if wallet.Balance != 0 {
		t.Errorf("期望钱包余额为 0，但得到 %v", wallet.Balance)
	}
}
