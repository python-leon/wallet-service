# Wallet Service

一个简单的数字钱包服务，支持创建钱包、查询余额和转账功能。

## 功能特性

- 创建新钱包
- 查询钱包信息
- 钱包间转账
- 内存存储（适用于开发和测试）

## 技术栈

- Go 1.22+
- 标准库 HTTP 服务器
- Go modules 依赖管理

## 项目结构

```
wallet-service/
├── api/
│   └── proto/          # API 协议定义
├── cmd/
│   └── server/         # 主程序入口
├── internal/
│   ├── errors/         # 自定义错误类型
│   ├── handler/        # HTTP 处理器
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
├── pkg/                # 可复用包
├── scripts/            # 脚本文件
└── test/               # 测试文件
```

## 快速开始

### 环境要求

- Go 1.22 或更高版本
- Git

### 安装和运行

1. 克隆项目：

```bash
git clone <repository-url>
cd wallet-service
```

2. 下载依赖：

```bash
go mod download
```

3. 运行服务：

**运行REST API服务：**
```bash
go run cmd/server/main.go
```

**运行gRPC服务：**
```bash
go run cmd/grpc_server/main.go
```

**运行统一服务（同时提供REST和gRPC）：**
```bash
go run cmd/unified_server/main.go
```

默认情况下，REST API服务将在 `http://localhost:8080` 上启动，gRPC服务在 `localhost:9090` 上启动。

### 构建二进制文件

```bash
# 构建REST API服务
go build -o bin/rest-server cmd/server/main.go

# 构建gRPC服务
go build -o bin/grpc-server cmd/grpc_server/main.go

# 构建统一服务
go build -o bin/unified-server cmd/unified_server/main.go
```

## Docker 支持

### 环境要求

- Docker 20.10+
- Docker Compose 1.29+（可选）

### 使用 Docker Compose（推荐）

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建镜像
docker-compose up -d --build
```

服务启动后：
- REST API: http://localhost:8080
- gRPC: localhost:50051

### 使用 Docker 直接构建

```bash
# 构建镜像
docker build -t wallet-service:latest .

# 运行容器
docker run -d \
  --name wallet-service \
  -p 8080:8080 \
  -p 50051:50051 \
  -e HTTP_PORT=8080 \
  -e GRPC_PORT=50051 \
  wallet-service:latest

# 查看日志
docker logs -f wallet-service

# 停止容器
docker stop wallet-service
docker rm wallet-service
```

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| HTTP_PORT | 8080 | REST API 端口 |
| GRPC_PORT | 50051 | gRPC 服务端口 |
| DEBUG | false | 调试模式 |

### 镜像特性

- **多阶段构建**：使用 `golang:1.24-alpine` 构建，`alpine:3.19` 运行，镜像体积小
- **安全性**：以非 root 用户运行
- **健康检查**：内置 HTTP 健康检查端点 `/health`
- **时区支持**：包含 tzdata，支持时区配置

### Docker 测试环境

启动包含 grpcurl 测试工具的环境：

```bash
# 启动服务（包含测试工具）
docker-compose --profile testing up -d

# 使用 grpcurl 测试
docker exec grpcurl-client grpcurl -plaintext wallet-service:50051 list
```

### 常见问题

**镜像拉取失败（国内网络）**

Dockerfile 已配置 Go 代理 `GOPROXY=https://goproxy.cn,direct`。如需配置 Docker 镜像加速，编辑 `/etc/docker/daemon.json`：

```json
{
  "registry-mirrors": [
    "https://docker.1ms.run"
  ]
}
```

然后重启 Docker：`sudo systemctl restart docker`

**端口被占用**

修改端口映射：
```bash
docker run -d -p 8081:8080 -p 50052:50051 wallet-service:latest
```

或通过环境变量：
```bash
docker run -d -p 8081:8081 -p 50052:50052 -e HTTP_PORT=8081 -e GRPC_PORT=50052 wallet-service:latest
```

## API 接口

### REST API

#### 创建钱包

```
POST /wallets
```

创建一个新的钱包，初始余额为 0。

**示例请求：**

```bash
curl -X POST http://localhost:8080/wallets
```

**示例响应：**

```json
{
  "success": true,
  "data": {
    "id": "wallet-id-here",
    "balance": 0,
    "created_at": "2026-03-13T10:00:00Z",
    "updated_at": "2026-03-13T10:00:00Z"
  }
}
```

#### 获取钱包信息

```
GET /wallets/{wallet_id}
```

根据钱包ID获取钱包信息。

**示例请求：**

```bash
curl http://localhost:8080/wallets/{1351dc5c-38c7-4c9d-addc-f69687eeb032}
```

**示例响应：**

```json
{
  "success": true,
  "data": {
    "id": "wallet-id-here",
    "balance": 100,
    "created_at": "2026-03-13T10:00:00Z",
    "updated_at": "2026-03-13T10:00:00Z"
  }
}
```

#### 存款

```
POST /wallets/deposit
```

为钱包充值。

**请求体：**

```json
{
  "wallet_id": "wallet-id",
  "amount": 1000
}
```

**示例请求：**

```bash
curl -X POST http://localhost:8080/wallets/deposit \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_id": "wallet-id-here",
    "amount": 1000
  }'
```

**示例响应：**

```json
{
  "success": true,
  "data": {
    "id": "wallet-id-here",
    "balance": 1000,
    "created_at": "2026-03-13T10:00:00Z",
    "updated_at": "2026-03-13T10:00:00Z"
  }
}
```

#### 转账

```
POST /wallets/transfer
```

在两个钱包之间转账。

**请求体：**

```json
{
  "from_wallet_id": "source-wallet-id",
  "to_wallet_id": "destination-wallet-id",
  "amount": 50
}
```

**示例请求：**

```bash
curl -X POST http://localhost:8080/wallets/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "from_wallet_id": "wallet1-id",
    "to_wallet_id": "wallet2-id",
    "amount": 50
  }'
```

**示例响应：**

```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "transfer successful",
    "from_wallet_id": "wallet1-id",
    "to_wallet_id": "wallet2-id",
    "from_balance": 50,
    "to_balance": 150,
    "transferred_at": "2026-03-13T10:00:00Z"
  }
}
```

## 测试

### 运行所有测试

```bash
go test ./...
```

### 运行特定包的测试

```bash
# 运行处理函数层测试
go test ./internal/handler

# 运行服务层测试
go test ./internal/service

# 运行仓库层测试
go test ./internal/repository
```

### 运行测试并查看详细输出

```bash
go test -v ./...
```

### 运行测试并生成覆盖率报告

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### gRPC 接口测试

项目提供了 gRPC 接口测试脚本，使用 `grpcurl` 工具进行测试。

#### 前置要求

安装 grpcurl：

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

#### 运行测试脚本

1. 首先启动服务：

```bash
go run cmd/server/main.go
```

2. 在另一个终端运行测试脚本：

```bash
./scripts/test_grpc.sh
```

#### 自定义配置

可以通过环境变量配置测试目标：

```bash
GRPC_HOST=localhost GRPC_PORT=50051 ./scripts/test_grpc.sh
```

#### 手动测试 gRPC 接口

使用 grpcurl 手动测试各个接口：

```bash
# 列出可用服务
grpcurl -plaintext localhost:50051 list

# 查看服务描述
grpcurl -plaintext localhost:50051 describe wallet.WalletService

# 创建钱包
grpcurl -plaintext -d '{}' localhost:50051 wallet.WalletService/CreateWallet

# 获取钱包信息
grpcurl -plaintext -d '{"wallet_id": "your-wallet-id"}' localhost:50051 wallet.WalletService/GetWallet

# 转账
grpcurl -plaintext -d '{
  "from_wallet_id": "wallet-1-id",
  "to_wallet_id": "wallet-2-id",
  "amount": 100
}' localhost:50051 wallet.WalletService/Transfer
```

### 负载测试

项目提供了负载测试工具，用于测试服务的并发性能和转账正确性。

#### 运行负载测试

```bash
# 启动服务
go run cmd/server/main.go

# 创建钱包测试
go run scripts/loadtest.go -c 20 -n 1000

# 混合测试
go run scripts/loadtest.go -test mixed -c 10 -d 30s

# 并发转账正确性测试
go run scripts/loadtest.go -test transfer -c 10 -n 100
```

详细的并发转账正确性测试说明请参阅 [docs/concurrency-test.md](docs/concurrency-test.md)。

## 开发指南

### 添加新功能

1. 在 `internal/model` 中定义数据结构
2. 在 `internal/repository` 中实现数据访问逻辑
3. 在 `internal/service` 中实现业务逻辑
4. 在 `internal/handler` 中实现HTTP处理逻辑
5. 编写相应的单元测试

### 代码规范

- 遵循 Go 语言规范和最佳实践
- 使用有意义的变量和函数命名
- 为公共函数和类型添加注释
- 编写单元测试覆盖核心功能

## 配置

服务可以通过环境变量进行配置：

- `PORT`: 服务监听端口，默认为 `8080`
- `DEBUG`: 是否启用调试模式，默认为 `false`

## 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 许可证

MIT License