.PHONY: dev build clean test help

# 默认目标
help: ## 显示帮助信息
	@echo CraftFire 构建命令:
	@echo   make dev        - 启动开发模式
	@echo   make build      - 构建生产版本
	@echo   make build-win  - 构建 Windows 版本
	@echo   make clean      - 清理构建缓存
	@echo   make test       - 运行所有测试
	@echo   make test-go    - 运行 Go 测试
	@echo   make test-fe    - 运行前端测试
	@echo   make lint       - 运行代码检查
	@echo   make install    - 安装依赖

# 开发
dev: ## 启动 Wails 开发模式
	cd backend && wails dev

# 构建
build: ## 构建生产版本
	cd backend && wails build

build-win: ## 构建 Windows 可执行文件
	cd backend && wails build --platform windows/amd64

build-nsis: ## 构建 Windows 安装程序
	cd backend && wails build -nsis

build-release: ## 优化构建（移除调试符号）
	cd backend && wails build -ldflags="-s -w"

# 清理
clean: ## 清理构建缓存
	cd backend && wails build -clean
	cd frontend && rm -rf node_modules dist
	cd backend && go clean -cache

# 测试
test: test-go test-fe ## 运行所有测试

test-go: ## 运行 Go 后端测试
	cd backend && go test ./... -v

test-go-cover: ## 运行 Go 测试（带覆盖率）
	cd backend && go test -cover ./... -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html

test-go-bench: ## 运行 Go 性能基准测试
	cd backend && go test -bench ./... -benchmem

test-fe: ## 运行前端单元测试
	cd frontend && npm run test:unit

# 安装依赖
install: ## 安装所有依赖
	cd frontend && npm install
	cd backend && go mod tidy

# 代码检查
lint: ## 运行代码检查
	cd backend && golangci-lint run ./...
	cd frontend && npm run lint
