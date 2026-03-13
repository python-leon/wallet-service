package grpc

import (
	"context"
	"time"

	"github.com/python-leon/wallet-service/internal/model"
	"github.com/python-leon/wallet-service/internal/service"
	pb "github.com/python-leon/wallet-service/proto"
)

// WalletGRPCServer 实现 gRPC 钱包服务
type WalletGRPCServer struct {
	pb.UnimplementedWalletServiceServer
	walletService service.WalletService
}

// NewWalletGRPCServer 创建新的 gRPC 服务器实例
func NewWalletGRPCServer(walletSvc service.WalletService) *WalletGRPCServer {
	return &WalletGRPCServer{
		walletService: walletSvc,
	}
}

// CreateWallet 创建新钱包
func (s *WalletGRPCServer) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	wallet := s.walletService.CreateWallet()

	resp := &pb.CreateWalletResponse{
		Success: true,
		Wallet: &pb.Wallet{
			Id:        wallet.ID,
			Balance:   wallet.Balance,
			CreatedAt: wallet.CreatedAt.Unix(),
			UpdatedAt: wallet.UpdatedAt.Unix(),
		},
	}

	return resp, nil
}

// GetWallet 获取钱包信息
func (s *WalletGRPCServer) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	wallet, err := s.walletService.GetWallet(req.WalletId)
	if err != nil {
		return &pb.GetWalletResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	resp := &pb.GetWalletResponse{
		Success: true,
		Wallet: &pb.Wallet{
			Id:        wallet.ID,
			Balance:   wallet.Balance,
			CreatedAt: wallet.CreatedAt.Unix(),
			UpdatedAt: wallet.UpdatedAt.Unix(),
		},
	}

	return resp, nil
}

// Transfer 执行转账操作
func (s *WalletGRPCServer) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	// 构造服务层需要的请求对象
	transferReq := &model.TransferRequest{
		FromWalletID: req.FromWalletId,
		ToWalletID:   req.ToWalletId,
		Amount:       req.Amount,
	}

	// 调用服务层的转账方法
	result, err := s.walletService.Transfer(transferReq)
	if err != nil {
		return nil, err
	}

	resp := &pb.TransferResponse{
		Success:       result.Success,
		Message:       result.Message,
		FromWalletId:  result.FromWalletID,
		ToWalletId:    result.ToWalletID,
		FromBalance:   result.FromBalance,
		ToBalance:     result.ToBalance,
		TransferredAt: time.Now().Unix(),
	}

	return resp, nil
}
