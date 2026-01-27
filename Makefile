.PHONY: all build run test clean deps

# 变量定义
APP_NAME=ai-agent-assistant
BUILD_DIR=bin
CMD_DIR=cmd/server
MAIN_FILE=$(CMD_DIR)/main.go

# 默认目标
all: deps build

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 构建
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run:
	@echo "Running $(APP_NAME)..."
	go run $(MAIN_FILE)

# 测试
test:
	@echo "Running tests..."
	go test -v ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Linting code..."
	golangci-lint run

# 清理
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	go clean

# 运行并热重载
dev:
	@echo "Running with hot reload..."
	air

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# Docker运行
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all          - Install dependencies and build (default)"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  dev          - Run with hot reload (requires air)"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  deps         - Install dependencies"
	@echo "  help         - Show this help message"
