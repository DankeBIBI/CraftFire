# CraftFire

> **Sandbox × FPS** 混合局域网多人游戏 · Low-Poly 风格

## 简介

CraftFire 是一款基于 Wails v2（Go 后端 + Vue 3 前端）的桌面游戏应用，融合了体素沙盒建造和第一人称射击玩法。游戏通过局域网（LAN）进行多人联机，无需外网服务器。

### 核心特色

- **体素建造** — 放置/移除石头、木头、玻璃等方块构建世界
- **FPS 战斗** — 第一人称视角射击，支持多种武器
- **局域网联机** — UDP 自动发现 + WebSocket 通信，6 位数字房间号
- **管理员面板** — 实时查看玩家列表、统计数据，支持踢出/禁言
- **3D 模型导入** — 支持 GLTF/GLB/FBX/OBJ/DAE 格式
- **Low-Poly 美术风格** — 全局一致的低多边形视觉设计

## 技术栈

| 层级     | 技术                                               |
| -------- | -------------------------------------------------- |
| 桌面框架 | Wails v2                                           |
| 后端     | Go 1.21+                                           |
| 前端     | Vue 3 (Composition API, Script Setup) + TypeScript |
| 3D 渲染  | TresJS v2 (Three.js 声明式封装)                    |
| 状态管理 | Pinia                                              |
| 样式     | Tailwind CSS + CSS Variables                       |
| 实时通信 | Gorilla WebSocket                                  |
| LAN 发现 | UDP 广播/监听                                      |
| 构建工具 | Vite 5                                             |

## 快速开始

```bash
# 克隆项目
git clone <repo-url>
cd CraftFire

# 安装前端依赖
cd frontend && npm install && cd ..

# 安装 Go 依赖
cd backend && go mod tidy && cd ..

# 启动开发模式
wails dev
```

## 构建

```bash
# Windows 可执行文件
wails build --output CraftFire.exe

# Windows NSIS 安装程序
wails build -nsis
```

## 项目结构

```
CraftFire/
├── frontend/          # Vue 3 前端
│   ├── src/
│   │   ├── components/   # Vue 组件
│   │   ├── stores/       # Pinia 状态管理
│   │   ├── services/     # 服务层
│   │   ├── composables/  # 组合式函数
│   │   ├── utils/        # 工具函数
│   │   └── types/        # TypeScript 类型
│   └── ...
├── backend/           # Go 后端
│   ├── internal/      # 内部模块
│   │   ├── websocket/ # WebSocket 通信
│   │   ├── room/      # 房间管理
│   │   ├── player/    # 玩家状态
│   │   ├── world/     # 体素世界
│   │   ├── lan/       # LAN 发现
│   │   ├── admin/     # 管理员面板
│   │   ├── model/     # 3D 模型管理
│   │   └── profile/   # 玩家资料
│   └── tests/         # 单元测试
├── docs/              # 文档
└── config.example.yml # 配置示例
```

## 文档

- [架构设计](docs/ARCHITECTURE.md)
- [安装指南](docs/INSTALL.md)
- [API 文档](docs/API.md)
- [贡献指南](docs/CONTRIBUTING.md)
- [故障排查](docs/TROUBLESHOOTING.md)

## License

MIT
