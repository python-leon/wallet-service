package main

import (
	"log"
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
	mux.HandleFunc("GET /wallets/{wallet_id}", walletHandler.GetWallet)
	mux.HandleFunc("POST /wallets/transfer", walletHandler.Transfer)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server
	log.Printf("Starting REST server on port %s...", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

}
