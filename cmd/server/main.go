package main

import (
	"log"
	"net"
	"net/http"
	"os"

	wg "github.com/python-leon/wallet-service/internal/grpc"
	"github.com/python-leon/wallet-service/internal/handler"
	"github.com/python-leon/wallet-service/internal/repository"
	"github.com/python-leon/wallet-service/internal/service"
	pb "github.com/python-leon/wallet-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize repository, service, and handler
	repo := repository.NewInMemoryWalletRepository()
	svc := service.NewWalletService(repo)

	// 启动 HTTP 服务
	walletHandler := handler.NewWalletHandler(svc)

	grpcHandler := wg.NewWalletGRPCServer(svc)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// Get ports from environment or use defaults
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Start gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterWalletServiceServer(grpcServer, grpcHandler)
		reflection.Register(grpcServer) // Enable gRPC reflection for tools like grpcurl

		log.Printf("Starting gRPC server on port %s...", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", walletHandler.CreateWallet)
	mux.HandleFunc("POST /wallets/transfer", walletHandler.Transfer)
	mux.HandleFunc("POST /wallets/deposit", walletHandler.Deposit)
	mux.HandleFunc("GET /wallets/{wallet_id}", walletHandler.GetWallet)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})

	// Start HTTP server
	log.Printf("Starting REST server on port %s...", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

}
