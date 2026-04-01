# CraftFire 架构设计

## 概览

CraftFire 采用 **Wails v2** 框架，将 Go 后端与 Vue 3 前端打包为单一桌面应用。

```
┌─────────────────────────────────────┐
│           Wails Desktop App          │
│  ┌───────────┐    ┌───────────────┐ │
│  │  Go 后端   │◄──►│  Vue 3 前端   │ │
│  │ (逻辑权威) │    │ (渲染 + UI)   │ │
│  └─────┬─────┘    └───────┬───────┘ │
│        │                  │         │
│        ▼                  ▼         │
│  ┌───────────┐    ┌───────────────┐ │
│  │ WebSocket  │    │   TresJS 3D   │ │
│  │   Server   │    │   Renderer    │ │
│  └───────────┘    └───────────────┘ │
└─────────────────────────────────────┘
         ▲
         │ LAN WebSocket
         ▼
    ┌──────────┐
    │ 其他玩家  │
    │ (客户端)  │
    └──────────┘
```

## 核心架构原则

### 服务端权威 (Server-Authoritative)

- Go 后端是游戏规则的唯一权威
- 客户端做**预测**（移动、放置），服务端做**校正**
- 8 TPS (每秒 8 次 Tick) 服务端更新频率
- 60 FPS 客户端渲染帧率，通过**插值**平滑 Tick 间隔

### 数据流

```
客户端输入 → 本地预测 → 发送到服务端 → 服务端校验 → 广播到所有客户端 → 客户端插值渲染
```

## 后端模块

| 模块        | 职责                                |
| ----------- | ----------------------------------- |
| `websocket` | WebSocket Hub/Client 管理、消息路由 |
| `room`      | 房间创建/销毁、玩家管理             |
| `player`    | 玩家状态、生命值、背包              |
| `world`     | 体素数据、区块管理、地形生成        |
| `lan`       | UDP 广播/发现局域网服务             |
| `admin`     | 管理员认证（JWT Token）、统计数据   |
| `model`     | 3D 模型导入、校验、LAN 同步         |
| `profile`   | 玩家资料、统计数据持久化            |
| `config`    | YAML 配置加载、环境变量覆盖         |
| `logger`    | 分级日志系统                        |

## 前端架构

### 组件层级

```
App.vue
├── MainMenu (菜单状态)
│   ├── RoomCreate
│   ├── RoomJoin
│   └── LANServerList
├── GameScene (游戏状态)
│   ├── TresCanvas (3D 场景)
│   │   ├── FirstPersonController
│   │   ├── VoxelWorld
│   │   ├── BlockPlacer
│   │   ├── PlayerEntity (本地)
│   │   └── RemotePlayer[] (远程)
│   └── GameHUD (HUD 叠加)
├── PauseMenu
├── SettingsPanel
├── AdminPanel
├── PlayerProfile
├── ModelImporter
└── Toast (全局通知)
```

### 状态管理 (Pinia Stores)

| Store       | 职责                              |
| ----------- | --------------------------------- |
| `gameState` | 游戏运行时状态、视图切换          |
| `websocket` | WS 连接状态、延迟统计             |
| `player`    | 本地/远程玩家数据                 |
| `room`      | 房间信息、LAN 服务器列表          |
| `world`     | 方块数据、区块管理                |
| `ui`        | Toast、弹窗、面板开关             |
| `settings`  | 用户偏好（持久化到 localStorage） |
| `admin`     | 管理员面板状态                    |
| `profile`   | 玩家个人资料                      |
| `model`     | 3D 模型管理                       |

## 通信协议

所有 WebSocket 消息使用 JSON Envelope 格式：

```json
{
  "type": "player_move",
  "timestamp": 1234567890,
  "playerId": "uuid",
  "roomId": "123456",
  "id": "msg-uuid",
  "payload": { ... }
}
```

### 消息类型

- `player_join` / `player_leave` — 玩家进出
- `player_move` — 位置/旋转更新
- `player_state_sync` — 批量状态同步
- `block_place` / `block_remove` — 方块操作
- `world_update` / `world_snapshot` — 世界数据
- `ping` / `pong` — 心跳检测

## 网络模型

- **传输**: WebSocket (TCP)
- **发现**: UDP 广播 (端口 9999)
- **房间号**: 6 位数字，同时作为 WebSocket 端口号
- **Tick Rate**: 8 TPS (125ms)
- **重连**: 指数退避，最多 5 次
