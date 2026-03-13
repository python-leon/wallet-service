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

```bash
go run cmd/server/main.go
```

默认情况下，服务将在 `http://localhost:8080` 上启动。

### 构建二进制文件

```bash
go build -o bin/wallet-service cmd/server/main.go
./bin/wallet-service
```

## API 接口

### 创建钱包

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

### 获取钱包信息

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

### 转账

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