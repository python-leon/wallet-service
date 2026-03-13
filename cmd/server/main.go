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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Get storage type from environment
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "memory" // default to in-memory storage
	}

	// Initialize repository based on storage type
	var repo repository.WalletRepository
	switch storageType {
	case "postgres":
		db, err := initDatabase()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		dbRepo := repository.NewDBWalletRepository(db)
		if err := dbRepo.AutoMigrate(); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		repo = dbRepo
		log.Println("Using PostgreSQL storage")
	default:
		repo = repository.NewInMemoryWalletRepository()
		log.Println("Using in-memory storage")
	}

	// Initialize service and handler
	svc := service.NewWalletService(repo)
	walletHandler := handler.NewWalletHandler(svc)
	grpcHandler := wg.NewWalletGRPCServer(svc)

	// Get ports from environment
	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50051")

	// Start gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterWalletServiceServer(grpcServer, grpcHandler)
		reflection.Register(grpcServer)

		log.Printf("Starting gRPC server on port %s...", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", walletHandler.CreateWallet)
	mux.HandleFunc("POST /wallets/transfer", walletHandler.Transfer)
	mux.HandleFunc("POST /wallets/deposit", walletHandler.Deposit)
	mux.HandleFunc("GET /wallets/{wallet_id}", walletHandler.GetWallet)
	mux.HandleFunc("GET /health", healthHandler)

	// Start HTTP server
	log.Printf("Starting REST server on port %s...", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func initDatabase() (*gorm.DB, error) {
	dsn := buildDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection established")
	return db, nil
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "wallet")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return "host=" + host + " port=" + port + " user=" + user +
		" password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
