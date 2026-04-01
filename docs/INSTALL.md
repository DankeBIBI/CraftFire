# CraftFire 安装指南

## 系统要求

### 必需

| 工具      | 版本                    |
| --------- | ----------------------- |
| Go        | ≥ 1.19（推荐 1.21+）    |
| Node.js   | ≥ 16.14（推荐 18 LTS+） |
| npm       | ≥ 8.x                   |
| Wails CLI | 最新版                  |

### 可选

- Git
- VS Code（推荐 IDE）

## 安装步骤

### 1. 安装 Go

从 [https://go.dev/dl/](https://go.dev/dl/) 下载并安装。

验证：

```bash
go version
# go version go1.21.x ...
```

### 2. 安装 Node.js

从 [https://nodejs.org/](https://nodejs.org/) 下载 LTS 版本。

验证：

```bash
node --version
# v18.x.x

npm --version
# 9.x.x
```

### 3. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

验证：

```bash
wails doctor
```

确保 `wails doctor` 输出没有错误。如果缺少依赖（如 WebView2、GCC 等），按提示安装。

### 4. 克隆项目

```bash
git clone <repo-url>
cd CraftFire
```

### 5. 安装依赖

```bash
# 前端
cd frontend
npm ci
cd ..

# 后端
cd backend
go mod tidy
cd ..
```

### 6. 启动开发

```bash
wails dev
```

首次启动可能需要下载额外依赖，请耐心等待。

## 构建发布版

### Windows

```bash
# 标准构建
wails build --output CraftFire.exe

# NSIS 安装程序
wails build -nsis

# 优化构建（去除调试符号）
wails build -ldflags="-s -w"
```

### 清理缓存

```bash
wails build -clean
```

## 配置

复制示例配置文件：

```bash
cp config.example.yml config.yml
```

编辑 `config.yml` 修改服务器端口、管理员密码等设置。

## VS Code 集成

项目已包含 `.vscode/` 配置：

- **推荐扩展**: 打开项目后 VS Code 会提示安装
- **调试配置**: F5 启动 Wails Dev 调试
- **设置**: Go 格式化、Vue 语法支持已预配置
