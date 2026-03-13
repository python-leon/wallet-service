package main

import (
	"net/http"
	"os"

	"github.com/python-leon/wallet-service/internal/handler"
	"github.com/python-leon/wallet-service/internal/repository"
	"github.com/python-leon/wallet-service/internal/service"
)

func main() {
	// Initialize repository, service, and handler
	repo := repository.NewInMemoryWalletRepository()
	svc := service.NewWalletService(repo)

	// 启动 HTTP 服务
	walletHandler := handler.NewWalletHandler(svc)

	// Get ports from environment or use defaults
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", walletHandler.CreateWallet)

}
