package handler

import (
	"encoding/json"
	"net/http"

	"github.com/python-leon/wallet-service/internal/model"
	"github.com/python-leon/wallet-service/internal/service"
	"github.com/python-leon/wallet-service/pkg/response"
)

type WalletHandler struct {
	service service.WalletService
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(svc service.WalletService) *WalletHandler {
	return &WalletHandler{
		service: svc,
	}
}

// CreateWallet handles POST /wallets
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	wallet := h.service.CreateWallet()
	response.Success(w, http.StatusCreated, wallet)
}

// GetWallet handles GET /wallets/{wallet_id}
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract wallet_id from URL path
	// Expected path: /wallets/{wallet_id}
	walletID := r.PathValue("wallet_id")
	if walletID == "" {
		response.Error(w, http.StatusBadRequest, "wallet_id is required")
		return
	}

	wallet, err := h.service.GetWallet(walletID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.Success(w, http.StatusOK, wallet)
}

// Transfer handles POST /wallets/transfer
func (h *WalletHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Transfer(&req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadRequest
	}

	response.JSON(w, statusCode, result)
}

// DepositRequest represents a deposit request
type DepositRequest struct {
	WalletID string `json:"wallet_id"`
	Amount   int64  `json:"amount"`
}

// Deposit handles POST /deposit
func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.WalletID == "" {
		response.Error(w, http.StatusBadRequest, "wallet_id is required")
		return
	}

	if req.Amount <= 0 {
		response.Error(w, http.StatusBadRequest, "amount must be positive")
		return
	}

	wallet, err := h.service.Deposit(req.WalletID, req.Amount)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, http.StatusOK, wallet)
}
