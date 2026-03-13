package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/python-leon/wallet-service/internal/model"
	"github.com/python-leon/wallet-service/internal/repository"
	"github.com/python-leon/wallet-service/internal/service"
	"github.com/python-leon/wallet-service/pkg/response"
)

// MockWalletService 是一个模拟的服务层实现，用于测试
type MockWalletService struct {
	CreateWalletFunc func() *model.Wallet
}

func (m *MockWalletService) CreateWallet() *model.Wallet {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc()
	}
	// 默认返回一个示例钱包
	return &model.Wallet{
		ID:      "test-wallet-id",
		Balance: 0,
	}
}

// 测试 CreateWallet 方法的成功场景
func TestCreateWallet_Success(t *testing.T) {
	// 创建模拟服务
	mockService := &MockWalletService{
		CreateWalletFunc: func() *model.Wallet {
			return &model.Wallet{
				ID:      "new-wallet-id",
				Balance: 0,
			}
		},
	}

	// 创建处理器实例
	handler := NewWalletHandler(mockService)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", "/wallets", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理函数
	handler.CreateWallet(rr, req)

	// 验证状态码
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("期望状态码 %v，但得到 %v", http.StatusCreated, status)
	}

	// 解析响应体
	var resp response.Response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应成功
	if !resp.Success {
		t.Errorf("期望响应成功，但得到失败")
	}

	// 验证响应数据
	walletData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Errorf("响应数据格式不正确")
	}

	if walletData["id"] != "new-wallet-id" {
		t.Errorf("期望钱包ID为 'new-wallet-id'，但得到 '%v'", walletData["id"])
	}

	if walletData["balance"].(float64) != 0 {
		t.Errorf("期望钱包余额为 0，但得到 %v", walletData["balance"])
	}
}

// 测试 CreateWallet 方法的错误场景 - 不正确的HTTP方法
func TestCreateWallet_MethodNotAllowed(t *testing.T) {
	// 创建模拟服务
	mockService := &MockWalletService{}

	// 创建处理器实例
	handler := NewWalletHandler(mockService)

	// 创建非POST请求
	req, err := http.NewRequest("GET", "/wallets", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理函数
	handler.CreateWallet(rr, req)

	// 验证状态码
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("期望状态码 %v，但得到 %v", http.StatusMethodNotAllowed, status)
	}

	// 解析响应体
	var resp response.Response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应失败
	if resp.Success {
		t.Errorf("期望响应失败，但得到成功")
	}

	if resp.Error != "method not allowed" {
		t.Errorf("期望错误消息为 'method not allowed'，但得到 '%v'", resp.Error)
	}
}

// 测试 CreateWallet 方法与真实服务集成
func TestCreateWallet_Integration(t *testing.T) {
	// 使用真实的仓库和服务层
	repo := repository.NewInMemoryWalletRepository()
	svc := service.NewWalletService(repo)
	handler := NewWalletHandler(svc)

	// 创建 POST 请求
	req, err := http.NewRequest("POST", "/wallets", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理函数
	handler.CreateWallet(rr, req)

	// 验证状态码
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("期望状态码 %v，但得到 %v", http.StatusCreated, status)
	}

	// 解析响应体
	var resp response.Response
	bodyBytes, _ := io.ReadAll(rr.Body)
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		t.Fatalf("解析响应失败: %v, 响应体: %s", err, string(bodyBytes))
	}

	// 验证响应成功
	if !resp.Success {
		t.Errorf("期望响应成功，但得到失败，响应: %+v", resp)
	}

	// 验证响应数据
	walletData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Errorf("响应数据格式不正确，类型: %T, 值: %+v", resp.Data, resp.Data)
	}

	// 验证钱包ID存在且不为空
	id, exists := walletData["id"]
	if !exists || id == "" {
		t.Errorf("期望钱包有ID，但ID不存在或为空")
	}

	// 验证余额为0
	balance, exists := walletData["balance"]
	if !exists || balance != float64(0) {
		t.Errorf("期望钱包余额为 0，但得到 %v", balance)
	}
}
