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
