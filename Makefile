# Makefile for Alert Management Platform

# 变量定义
APP_NAME := alert-management-platform
BIN_DIR := bin
CMD_DIR := cmd/alert-management-platform
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_DEV_IMAGE := $(APP_NAME):dev
GO_VERSION := 1.21

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 构建相关
.PHONY: build
build: ## 构建应用程序
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "Build completed: $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-local
build-local: ## 构建本地版本
	@echo "Building $(APP_NAME) for local..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "Local build completed: $(BIN_DIR)/$(APP_NAME)"

.PHONY: clean
clean: ## 清理构建文件
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -rf tmp
	@go clean
	@echo "Clean completed"

# 测试相关
.PHONY: test
test: ## 运行测试
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "Running race tests..."
	@go test -race -v ./...

# 代码质量
.PHONY: fmt
fmt: ## 格式化代码
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet: ## 运行go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: lint
lint: ## 运行golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

.PHONY: check
check: fmt vet test ## 运行所有检查（格式化、vet、测试）

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "Downloading dependencies..."
	@go mod download

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

.PHONY: deps-verify
deps-verify: ## 验证依赖
	@echo "Verifying dependencies..."
	@go mod verify

# 运行相关
.PHONY: run
run: ## 运行应用程序
	@echo "Running $(APP_NAME)..."
	@go run ./$(CMD_DIR)

.PHONY: run-dev
run-dev: ## 使用air运行开发模式
	@echo "Running $(APP_NAME) in development mode..."
	@air -c .air.toml

# Docker相关
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

.PHONY: docker-build-dev
docker-build-dev: ## 构建开发Docker镜像
	@echo "Building development Docker image..."
	@docker build -f Dockerfile.dev -t $(DOCKER_DEV_IMAGE) .
	@echo "Development Docker image built: $(DOCKER_DEV_IMAGE)"

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --rm $(DOCKER_IMAGE)

.PHONY: docker-compose-up
docker-compose-up: ## 启动docker-compose服务
	@echo "Starting docker-compose services..."
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## 停止docker-compose服务
	@echo "Stopping docker-compose services..."
	@docker-compose down

.PHONY: docker-compose-dev-up
docker-compose-dev-up: ## 启动开发环境docker-compose服务
	@echo "Starting development docker-compose services..."
	@docker-compose -f docker-compose.dev.yml up -d

.PHONY: docker-compose-dev-down
docker-compose-dev-down: ## 停止开发环境docker-compose服务
	@echo "Stopping development docker-compose services..."
	@docker-compose -f docker-compose.dev.yml down

.PHONY: docker-logs
docker-logs: ## 查看docker-compose日志
	@docker-compose logs -f

.PHONY: docker-clean
docker-clean: ## 清理Docker资源
	@echo "Cleaning Docker resources..."
	@docker system prune -f
	@docker volume prune -f

# 数据库相关
.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	@echo "Running database migrations..."
	@go run ./scripts/migrate.go

.PHONY: db-seed
db-seed: ## 填充测试数据
	@echo "Seeding database..."
	@go run ./scripts/seed.go

.PHONY: db-reset
db-reset: ## 重置数据库
	@echo "Resetting database..."
	@go run ./scripts/reset.go

# 安装工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"

# 生成相关
.PHONY: generate
generate: ## 运行go generate
	@echo "Running go generate..."
	@go generate ./...

# 发布相关
.PHONY: release
release: clean test build ## 构建发布版本
	@echo "Release build completed"

# 健康检查
.PHONY: health
health: ## 检查应用健康状态
	@echo "Checking application health..."
	@curl -f http://localhost:8080/health || echo "Application is not running"

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "Go version: $(shell go version)"
	@echo "App name: $(APP_NAME)"
	@echo "Docker image: $(DOCKER_IMAGE)"