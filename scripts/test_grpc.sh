#!/bin/bash

# gRPC 测试脚本
# 使用 grpcurl 测试 Wallet Service 的 gRPC 接口
# 
# 前置要求:
#   1. 安装 grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
#   2. 启动服务: go run cmd/server/main.go

set -e

# 配置
GRPC_HOST="${GRPC_HOST:-localhost}"
GRPC_PORT="${GRPC_PORT:-50051}"
GRPC_ADDR="${GRPC_HOST}:${GRPC_PORT}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印测试标题
print_test() {
    echo -e "\n${YELLOW}========================================${NC}"
    echo -e "${YELLOW}测试: $1${NC}"
    echo -e "${YELLOW}========================================${NC}"
}

# 打印成功信息
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 打印错误信息
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# 检查 grpcurl 是否安装
check_grpcurl() {
    if ! command -v grpcurl &> /dev/null; then
        print_error "grpcurl 未安装"
        echo "请运行: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
        exit 1
    fi
    print_success "grpcurl 已安装"
}

# 检查服务是否运行
check_service() {
    print_test "检查服务状态"
    if grpcurl -plaintext "${GRPC_ADDR}" list &> /dev/null; then
        print_success "gRPC 服务运行在 ${GRPC_ADDR}"
    else
        print_error "无法连接到 gRPC 服务 ${GRPC_ADDR}"
        echo "请确保服务已启动: go run cmd/server/main.go"
        exit 1
    fi
}

# 测试列出服务
test_list_services() {
    print_test "列出可用服务"
    echo "grpcurl -plaintext ${GRPC_ADDR} list"
    grpcurl -plaintext "${GRPC_ADDR}" list
    print_success "服务列表获取成功"
}

# 测试服务描述
test_describe_service() {
    print_test "获取服务描述"
    echo "grpcurl -plaintext ${GRPC_ADDR} describe wallet.WalletService"
    grpcurl -plaintext "${GRPC_ADDR}" describe wallet.WalletService
    print_success "服务描述获取成功"
}

# 测试创建钱包
test_create_wallet() {
    print_test "创建钱包"
    echo "grpcurl -plaintext -d '{}' ${GRPC_ADDR} wallet.WalletService/CreateWallet"
    WALLET_RESPONSE=$(grpcurl -plaintext -d '{}' "${GRPC_ADDR}" wallet.WalletService/CreateWallet)
    echo "${WALLET_RESPONSE}"
    
    # 提取钱包 ID
    WALLET_ID=$(echo "${WALLET_RESPONSE}" | grep -o '"id": *"[^"]*"' | sed 's/"id": *"\([^"]*\)"/\1/')
    
    if [ -n "${WALLET_ID}" ]; then
        print_success "钱包创建成功, ID: ${WALLET_ID}"
        echo "${WALLET_ID}"
    else
        print_error "钱包创建失败"
        return 1
    fi
}

# 测试获取钱包
test_get_wallet() {
    local wallet_id="$1"
    print_test "获取钱包信息"
    echo "grpcurl -plaintext -d '{\"wallet_id\": \"${wallet_id}\"}' ${GRPC_ADDR} wallet.WalletService/GetWallet"
    grpcurl -plaintext -d "{\"wallet_id\": \"${wallet_id}\"}" "${GRPC_ADDR}" wallet.WalletService/GetWallet
    print_success "钱包信息获取成功"
}

# 测试转账
test_transfer() {
    local from_wallet_id="$1"
    local to_wallet_id="$2"
    local amount="$3"
    
    print_test "转账测试 (金额: ${amount})"
    echo "grpcurl -plaintext -d '{\"from_wallet_id\": \"${from_wallet_id}\", \"to_wallet_id\": \"${to_wallet_id}\", \"amount\": ${amount}}' ${GRPC_ADDR} wallet.WalletService/Transfer"
    grpcurl -plaintext -d "{\"from_wallet_id\": \"${from_wallet_id}\", \"to_wallet_id\": \"${to_wallet_id}\", \"amount\": ${amount}}" "${GRPC_ADDR}" wallet.WalletService/Transfer
}

# 测试错误处理 - 获取不存在的钱包
test_get_nonexistent_wallet() {
    print_test "获取不存在的钱包 (错误处理测试)"
    echo "grpcurl -plaintext -d '{\"wallet_id\": \"non-existent-id\"}' ${GRPC_ADDR} wallet.WalletService/GetWallet"
    grpcurl -plaintext -d '{"wallet_id": "non-existent-id"}' "${GRPC_ADDR}" wallet.WalletService/GetWallet
    print_success "错误处理测试完成"
}

# 主测试流程
main() {
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}   Wallet Service gRPC 测试脚本${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo "目标服务: ${GRPC_ADDR}"
    
    # 检查环境
    check_grpcurl
    check_service
    
    # 运行测试
    test_list_services
    test_describe_service
    
    # 创建两个钱包用于测试
    print_test "创建测试钱包"
    WALLET1_ID=$(test_create_wallet | tail -1)
    WALLET2_ID=$(test_create_wallet | tail -1)
    
    # 获取钱包信息
    test_get_wallet "${WALLET1_ID}"
    test_get_wallet "${WALLET2_ID}"
    
    # 测试错误处理
    test_get_nonexistent_wallet
    
    # 测试转账 (会失败，因为余额不足，但可以验证接口调用)
    test_transfer "${WALLET1_ID}" "${WALLET2_ID}" 10
    
    echo -e "\n${GREEN}======================================${NC}"
    echo -e "${GREEN}   所有测试完成!${NC}"
    echo -e "${GREEN}======================================${NC}"
}

# 运行主函数
main "$@"