package handler

import (
	"net/http"

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
