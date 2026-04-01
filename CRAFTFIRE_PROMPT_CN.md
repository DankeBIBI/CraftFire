# 项目："CraftFire" (沙盒 + FPS 混合游戏) - 基于 Wails & Vue 3 实现

你是一位精通 **Go (Golang)**、**Wails v2**、**Vue 3**、**TypeScript** 和 **3D 图形学** 的全栈游戏开发专家。

我需要你构建并搭建一个结合了沙盒机制（类 Minecraft）与 FPS 元素（类 CrossFire/穿越火线）的桌面游戏项目。

## 1. 技术栈与库 (最佳实践)

请使用以下现代流行的库，以确保项目的可维护性和高性能：

### 后端 (容器与逻辑)

- **框架**: **Wails v2** (Go) - 用于原生应用打包和后端逻辑处理。
- **网络**: **Gorilla WebSocket** (`github.com/gorilla/websocket`) - 用于稳定健壮的本地实时通信。
- **服务发现**: 原生 Go `net` 包 - 用于 UDP 广播 (局域网发现)。

### 前端 (视图与渲染)

- **核心**: **Vue 3** (Script Setup, 组合式 API) + **TypeScript** + **Vite**。
- **3D 引擎**: **TresJS** (v2) - Vue 生态中最流行的声明式 Three.js 封装库。它允许你像写 Vue 组件一样使用 Three.js。
- **物理引擎 (推荐)**: **@tresjs/cannon** (Cannon-es 封装) - 用于重力和碰撞检测。
- **状态管理**: **Pinia** - 用于管理游戏状态（玩家坐标、背包、房间状态）。
- **样式**: **Tailwind CSS** - 用于 UI 界面（HUD、菜单）。
- **工具库**: **VueUse** - 用于输入处理（键盘/鼠标捕捉）。

## 2. 架构概览

应用程序遵循严格的解耦架构：

- **Wails (Go)**: 充当“服务端”和“容器”。它处理物理计算（可选），管理 WebSocket 服务器，执行局域网发现，并处理操作系统级的文件操作。
- **Vue 3 (Frontend)**: 充当“客户端”。它纯粹用于渲染由 Go 提供的数据状态，并捕获用户输入。

## 3. 详细需求

### A. Wails 后端 (“大脑”)

1. **WebSocket 服务器**:
   - 游戏启动时在本地托管一个 WebSocket 服务器。
   - 向连接的客户端广播游戏状态（玩家坐标、世界变化）。
   - **Tick Rate (刷新率)**: 固定为 **8 ticks/秒** (约每 125ms 更新一次)。
   - **端口规则**: 房间号为 6 位数字，创建房间时使用房间号作为 WS 端口号（例如房间号 `123456` -> 端口 `123456`），仅用于局域网内通信。
2. **局域网发现**:
   - 实现一种机制（可能是 UDP 组播）来搜索局域网内运行该游戏的其他设备。
   - 向前端暴露一个 `FindLANServers()` 方法，返回可用房间号与 IP 列表。
3. **房间逻辑**:
   - **创建房间**: 生成 6 位随机房间号，启动对应端口的 WS 服务。
   - **加入房间**: 校验 6 位房间号，连接指定 IP + 端口的 WS 服务。
   - 提供 `CreateRoom()` 与 `JoinRoom(roomId, ip)` 两个方法供前端调用。
4. **应用生命周期**:
   - 支持打包为独立的 `.exe` (Windows) 或带有安装程序的二进制文件。

### B. 前端 Web (“眼睛”)

1. **3D 世界 (TresJS)**:
   - 在 Wails webview 中渲染 3D 场景。
   - 支持导入模型：
     - **角色/武器** (GLTF/GLB)。
     - **地图/地形** (基于体素或网格)。
   - **摄像机**: 第一人称视角控制 (Pointer Lock)。
2. **沙盒交互**:
   - 实现射线检测 (Raycasting) 以检测物体。
   - **操作**: 创建 (放置方块)、销毁 (移除方块)、移动、旋转、缩放物体。
3. **同步系统**:
   - 连接到 Wails WebSocket 服务器。
   - 接收玩家坐标和世界数据。
   - 实现 **插值 (Interpolation)** 平滑算法，以处理 8-tick 的低刷新率，确保由更新间隔带来的移动视觉效果平滑流畅。
4. **房间 UI 与流程**:
   - 提供 **创建房间** 与 **加入房间** 两种入口（Web 端与 Wails 内置 Webview 端保持一致）。
   - **创建房间**: 前端触发创建，显示生成的 6 位房间号，并在界面显著位置展示。
   - **加入房间**: 用户输入 6 位房间号进行加入，支持复制/粘贴。
   - 如果是局域网发现到的房间，允许一键加入。
5. **3D 文件导入与渲染系统**:

   - **文件格式支持**: 支持导入流行的 3D 模型格式：
     - **GLTF/GLB** (推荐) - 最佳兼容性，TresJS 原生支持
     - **FBX** (需转换) - 通过预处理工具转换为 GLTF
     - **OBJ/MTL** - 简单模型支持
     - **Collada (DAE)** - 可选
   - **导入流程**:
     - Web 端提供"导入 3D 模型"按钮，打开文件选择器
     - 支持拖拽上传 3D 文件到窗口中
     - 前端验证文件格式与大小（建议单个文件 < 10MB）
     - 上传到本地 Go 后端临时存储目录
   - **局域网模型共享**:
     - 房间内的玩家登录时自动检测本地已下载的 3D 模型列表
     - 后端维护模型目录 (`~/.CraftFire/models/`)
     - 房间启动时通过 WebSocket 广播可用模型清单
     - 局域网用户因需求可自动拉取下载缺失的模型，无需手动上传
     - 使用 MD5/SHA256 校验模型文件完整性，避免重复下载
     - 支持增量下载与断点续传
   - **渲染与集成**:
     - 导入的模型在场景中实时渲染（使用 TresJS）
     - 支持模型材质、纹理、动画保留
     - 可设置模型缩放、旋转、位置（通过 Inspector 或快捷菜单）
     - 模型可作为装饰、武器或环境元素
     - 支持模型的物理碰撞（通过 @tresjs/cannon）

6. **个人页面 - 角色形象预览**:

   - **页面入口**: 在游戏 HUD 右上角或主菜单导航栏加入"个人页面"按钮
   - **功能模块**:
     - **3D 角色模型预览区**:
       - 中央显示当前玩家的 3D 角色模型（默认或自定义）
       - 支持鼠标拖拽实时 360° 旋转查看
       - 支持缩放（鼠标滚轮）与平移（右键拖拽）
       - 显示模型装备、护甲、武器位置
       - 支持播放角色动画（待机、行走、攻击等动画切换）
     - **玩家信息展示区**:
       - 昵称、ID、加入时间、在线时长
       - 当前等级、经验进度条
       - 游戏统计：击杀数、死亡数、建造方块数、销毁方块数等
     - **装备与背包查看**:
       - 装备栏：当前武器、护甲、饰品
       - 背包物品列表（网格布局，可滚动）
     - **成就面板** (可选):
       - 显示已解锁的成就徽章
       - 未完成成就的进度显示
     - **设置与定制** (可选):
       - 修改角色昵称
       - 选择角色模型皮肤（从已有资产库选择）
       - 角色颜色/装扮定制
   - **技术实现**:
     - 使用 TresJS 独立场景渲染角色模型（与游戏场景分离）
     - 配置固定摄像机与灯光以实现最佳预览效果
     - 响应式布局：适配不同屏幕尺寸
     - 个人信息数据从 Go 后端通过 WebSocket 或 REST API 获取

7. **整体 UI 风格**:
   - 前端整体视觉与交互风格为 **Low Poly**（低多边形）风格。
   - 3D 世界中的模型、地形、角色与武器优先使用 Low Poly 资产。
   - UI 视觉元素保持棱角、几何感、低面数阴影与块面色彩的统一感。

### 7.1 Tailwind 实现 — Low‑Poly 多色设计规范（推荐）

- 目标：使用 Tailwind 的 design‑tokens 与原子类构建一致且可定制的 Low‑Poly 多色主题，便于在 monorepo 中复用与按需覆盖。

#### 设计要点（一句话）

- 使用有限的色板（主/次/强调/背景/文本），用硬边阴影与 1–2px 的实体边框强化“像素/低多边形”质感；通过 token 命名（`craft-*`）保持语义化。

#### 建议的 Design‑tokens（示例）

- 颜色：
  - craft-bg: #0b1220（深）
  - craft-surface: #0f1a2a（面）
  - craft-text: #e6eef8（主文案）
  - craft-primary: #00d1b2（主色，亮）
  - craft-secondary: #ffd166（次色）
  - craft-accent: #ff6b6b（警示/危险）
- 阴影：`shadow-lowpoly`（硬边投影）
- 字体/像素化：`font-family: 'Press Start 2P'`（可做可选 fallback）

#### 推荐的 `tailwind.config` 片段（放在项目根或 workspace shared styles）

```js
// theme.extend 的关键项（示例）
module.exports = {
  theme: {
    extend: {
      colors: {
        'craft-dark': '#0b1220',
        'craft-surface': '#0f1a2a',
        'craft-light': '#e6eef8',
        'craft-primary': '#00d1b2',
        'craft-secondary': '#ffd166',
        'craft-accent': '#ff6b6b',
      },
      boxShadow: {
        lowpoly: '6px 6px 0 rgba(0,0,0,0.85)',
        'lowpoly-lg': '10px 10px 0 rgba(0,0,0,0.9)',
      },
      fontFamily: {
        game: ['"Press Start 2P"', 'monospace'],
      },
      borderRadius: { sm: '2px' },
    },
  },
  safelist: ['btn-lowpoly', 'panel-lowpoly', 'input-lowpoly'],
};
```

#### 可复用组件类（放到 `@layer components` 并用 `@apply`）

- `btn-lowpoly`: px-6 py-3 font-game text-sm uppercase tracking-wider border-2 border-black shadow-lowpoly active:translate-x-1 active:translate-y-1
- `btn-primary`: `@apply btn-lowpoly bg-craft-primary text-white hover:brightness-105`
- `input-lowpoly`: `@apply w-full px-4 py-3 font-game text-sm bg-craft-surface text-craft-light border-2 border-craft-primary shadow-lowpoly focus:border-craft-secondary`
- `panel-lowpoly`: `@apply bg-craft-dark/90 backdrop-blur-sm border-2 border-craft-primary shadow-lowpoly-lg p-6`

（在 `packages/styles` 或 `apps/*/src/styles` 中维护一份共享的 `craft-ui.css`）

#### 示例：在 Vue 组件中的使用

```html
<button class="btn-lowpoly btn-primary">开始</button>
<input class="input-lowpoly" placeholder="搜索物品..." />
<section class="panel-lowpoly">面板内容</section>
```

#### 可访问性 & 设备回退

- 强制提供 `:focus-visible` 可见样式（不要仅依赖 :focus 的颜色改变）。
- 支持 `prefers-reduced-motion` 与 `prefers-reduced-transparency` 回退。
- 滚动条为装饰性实现（仅 WebKit），需要为非 WebKit 提供可见替代或隐藏策略。

#### 多色主题与可定制性

- 使用 CSS 变量 + Tailwind tokens（例如 `--craft-primary` → `text-craft-primary`），方便运行时主题切换与按房间/地图着色（如：industrial / forest / neon 预设）。
- 提供 3 个预设主题：`craft-default`、`craft-neon`、`craft-earth`（在 `tailwind.config` 中用 `variants` 或通过 CSS 变量切换）。

#### 开发与迁移建议

1. 在共享包（如 `packages/styles` 或 `@core/styles`）中维护 `craft-ui.css` 与 `tailwind.config` 的 tokens。✅
2. 逐步替换全局 CSS：先把 design‑tokens 与变量迁移到 Tailwind，再把组件样式用 `@apply` 封装成语义类，最后用原子类精细化样式。✅
3. 为动态类（根据状态生成的类名）在 `tailwind.config.safelist` 中列出避免被 purge 丢失。✅

#### PR 模板检查项（建议）

- [ ] 是否在 `tailwind.config` 中新增/复用 token？
- [ ] 是否在共享样式包新增了 `@layer components` 的类？
- [ ] 是否包含 `:focus-visible` 与 `prefers-reduced-*` 回退？
- [ ] 是否更新了 Storybook / demo（若有）以展示新主题？

> 结论：在提示词中加入上述 Tailwind 实践能让设计规范更可执行 —— 开发者会得到一套可复用的 token、组件类与迁移路径，有助于在 monorepo 多个 app 中保持视觉一致性。

8. **管理员面板（可选）**:
   - 在首页主菜单提供 **"管理员入口"** 按钮。
   - 要求输入指定的管理员密码进行身份验证。
   - 验证成功后进入管理员面板，显示当前房间内所有在线玩家的列表。
   - 玩家列表以表格形式排列，显示玩家名称、ID、状态、血量、位置等关键信息。
   - 点击列表中的玩家可弹出详情窗口，显示该玩家的完整信息（包括装备、背包、连接质量等）。
   - 管理员可进行踢除玩家、禁言、观察玩家等操作（可选）。

## 4. 代码规范

- **组件化**: 将 UI 和 3D 场景拆分为细小的、可复用的 Vue 组件 (例如 `<PlayerEntity />`, `<VoxelWorld />`, `<GameHUD />`)。
- **整洁代码**: 严格的 TypeScript 类型定义。为所有数据结构使用接口 (例如 `interface PlayerState { x: number; y: number; z: number; id: string; }`)。
- **注释**: 为所有主要函数添加 JSDoc/GoDoc 注释，且 **注释内容必须为中文**，解释输入/输出和用途。
- **拒绝面条代码**: 将渲染逻辑 (Vue) 与网络逻辑 (Pinia actions) 分离。

## 5. 数据持久化与状态管理

### A. 游戏状态存储

- **本地持久化**: 使用 Go 的文件系统存储房间配置、玩家建造データ、世界快照（使用 JSON 或 Gob 格式）。
- **内存态板**: 每个房间在 Go 后端维护内存中的实时游戏状态（玩家列表、方块网格、物体位置）。
- **初始化流程**: 房间启动时从本地加载上次保存的世界状态，服务器关闭前持久化当前状态。

### B. 数据同步策略

- **差量更新**: 只在客户端发送变化的部分（例如新放置的方块、移除的对象），避免每 tick 都发送完整世界状态。
- **确认机制**: 重要操作（放置/移除方块）需要客户端与服务器确认，防止不一致。

## 6. 安全性与数据校验

- **输入校验**: 所有来自前端的命令（放置方块、移动、攻击）在 Go 后端进行校验，检查：
  - 坐标有效性（范围检查、精度检查）。
  - 玩家权限与状态（是否在对应房间、是否存活）。
  - 操作频率限制（防止刷屏，如每秒最多放置 N 个方块）。
- **通信加密**: 考虑使用 TLS 对 WebSocket 连接加密（仅局域网环境，可选）。
- **房间隔离**: 确保房间之间数据隔离，一个房间的玩家无法访问其他房间的状态。

## 7. 错误处理与日志

### A. 错误处理策略

- **Go 后端**: 使用 Go 标准错误处理（`error` 接口），关键路径使用 `fmt.Errorf` 加上上下文。
- **前端**: 使用 TypeScript `try-catch` 捕获异步操作异常，并展示用户友好的错误提示（通过 Pinia store 的全局通知系统）。
- **WebSocket 异常**: 连接断开时自动重连（指数退避策略，最多重试 5 次），并提示用户"连接已断开"。

### B. 日志规范

- **Go**: 使用 `log` 标准库或第三方库（如 `logrus`）记录关键事件（房间创建/销毁、玩家加入/离开、异常错误）。
- **前端**: 在浏览器 DevTools Console 中记录关键状态变化和网络事件，使用 `console.log/warn/error` 配合前缀标识（如 `[GameState]`、`[Network]`）。

## 8. 性能优化指标

- **目标 FPS**: 前端渲染目标帧率 **60 FPS**，使用 `requestAnimationFrame` 控制。
- **网络延迟**: WebSocket 消息往返延迟应 < **200ms**（局域网环境下）。
- **内存占用**: 单房间状态内存占用 < **100MB**，支持最多 **50 个并发玩家**。
- **优化点**:
  - 使用对象池（Object Pool）复用方块和粒子效果对象。
  - 实现视锥剔除（Frustum Culling）减少不可见对象的渲染。
  - 使用 LOD（细节等级）动态调整远处模型复杂度。

## 9. API 接口定义

### WebSocket 消息格式

所有消息使用 JSON 格式，结构如下：

```json
{
  "type": "message_type",
  "timestamp": 1234567890,
  "data": {
    /* 具体数据 */
  }
}
```

### 关键消息类型

- **`player_move`**: 玩家位置更新 `{x, y, z, rotation}`。
- **`block_place`**: 放置方块 `{x, y, z, type}`。
- **`block_remove`**: 移除方块 `{x, y, z}`。
- **`player_spawn`**: 玩家加入房间 `{playerId, name, position}`。
- **`player_leave`**: 玩家离开房间 `{playerId}`。
- **`world_state`**: 同步世界完整状态（启动或大更新时）`{blocks, players}`。

### Go 后端暴露的 Wails 方法

**基础操作:**

- `CreateRoom() -> (roomId: string, error)`: 创建房间，返回 6 位房间号。
- `JoinRoom(roomId: string, playerName: string) -> (success: bool, error)`: 加入房间。
- `LeaveRoom() -> error`: 离开当前房间。
- `FindLANServers() -> (servers: [{roomId, ip, playerCount}], error)`: 发现局域网内的房间。
- `PlaceBlock(x, y, z, type: string) -> error`: 放置方块。
- `RemoveBlock(x, y, z) -> error`: 移除方块。

**管理员操作:**

- `VerifyAdminPassword(roomId: string, password: string) -> (isValid: bool, error)`: 验证管理员密码，返回是否合法。
- `GetOnlinePlayers(roomId: string) -> (players: [PlayerInfo], error)`: 获取房间内所有在线玩家列表。
- `GetPlayerDetails(roomId: string, playerId: string) -> (details: PlayerDetails, error)`: 获取指定玩家的详细信息。
- `KickPlayer(roomId: string, playerId: string, reason?: string) -> error`: 踢除玩家（可选）。
- `MutePlayer(roomId: string, playerId: string, durationSeconds: number) -> error`: 禁言玩家（可选）。
- `GetRoomStats(roomId: string) -> (stats: RoomStatistics, error)`: 获取房间统计数据。

**3D 模型导入与管理操作:**

- `ImportModel(filePath: string, roomId?: string) -> (modelId: string, error)`: 导入 3D 模型文件，返回模型 ID。
- `ListAvailableModels(roomId?: string) -> (models: [ModelInfo], error)`: 获取可用的 3D 模型列表。
- `GetModelInfo(modelId: string) -> (info: ModelInfo, error)`: 获取指定模型的详细信息（大小、格式、MD5、缩略图等）。
- `DeleteModel(modelId: string) -> error`: 删除已导入的 3D 模型。
- `SyncModelsInLAN() -> (changedModels: [ModelInfo], error)`: 同步局域网内所有可用的模型，返回新增/更新的模型列表。
- `DownloadModelFromLAN(modelId: string, sourceIP: string) -> error`: 从局域网其他设备下载指定模型。
- `GetPlayerProfile(playerId?: string) -> (profile: PlayerProfile, error)`: 获取玩家详细信息（如不指定 playerId 则获取当前玩家）。
- `UpdatePlayerProfile(profile: PlayerProfileUpdate) -> error`: 更新玩家个人信息（昵称、个性装扮等）。
- `GetPlayerStatistics(playerId?: string) -> (stats: PlayerStatistics, error)`: 获取玩家游戏统计数据。

## 10. 测试策略

### A. 后端测试 (Go)

- **单元测试**: 为 WebSocket Hub、房间管理、数据校验编写单元测试（使用 `testing` 包）。
- **集成测试**: 测试完整的房间创建、加入、玩家同步流程。
- **性能测试**: 压力测试房间并发连接能力（Benchmark）。

### B. 前端测试 (Vue)

- **组件测试**: 使用 `vitest` 或 `jest` 测试 Vue 组件（房间选择、HUD 显示）。
- **集成测试**: 测试 Pinia store 与 WebSocket 的交互。
- **E2E 测试**: 使用 Playwright 或 Cypress 测试完整的创建房间→加入游戏→交互流程。

## 11. 构建与部署

### A. 开发环境

- **前端开发**: `npm run dev` 启动 Vite 开发服务器（热更新）。
- **后端开发**: 使用 `wails dev` 启动 Wails 开发模式（自动重编译 Go 代码，同时运行前端服务）。

### B. 生产构建

- **编译**: 运行 `wails build` 生成平台原生二进制文件：
  - Windows: `CraftFire.exe`（可选配合 NSIS 安装程序）。
  - macOS: `CraftFire.app`。
  - Linux: `CraftFire` 二进制。
- **代码优化**: 启用编译器优化（Go ldflags），移除调试符号，减小二进制体积。

### C. 版本管理

- 使用语义化版本 (Semantic Versioning): `MAJOR.MINOR.PATCH`，例如 `1.0.0`。
- 在 `go.mod`、`package.json` 中维护版本号。
- 发布时标记 Git tag，例如 `v1.0.0`。

## 12. GIT 工作流程

- **分支策略**: 使用 GitFlow 或 Trunk-Based 开发（根据团队偏好）。
  - `main`: 稳定版本分支，只接受带标签的 release。
  - `dev`: 开发分支，集成所有功能分支。
  - `feature/*`: 功能开发分支，例如 `feature/voxel-rendering`。
- **提交规范**: 遵循 Conventional Commits，例如：
  - `feat(game): 实现第一人称摄像机控制`。
  - `fix(network): 修复 WebSocket 连接断开后的重连逻辑`。
  - `docs(api): 补充 WebSocket 消息格式文档`。
- **代码审查**: Pull Request 需要至少 1 人审核通过后才能合并。

## 13. 文档要求

### A. 项目文档

- **README.md**: 项目简介、快速开始、功能列表、项目结构。
- **ARCHITECTURE.md**: 详细的架构设计文档，包括数据流、模块交互、设计决策。
- **CONTRIBUTING.md**: 贡献指南，包括开发环境搭建、代码规范、提交流程。

### B. API 文档

- **WebSocket API**: 详细列出所有消息类型、参数、返回值示例（可使用 OpenAPI 格式或 Markdown）。
- **Go 后端 API**: 为输出的 Wails 方法生成 GoDoc 文档。

### C. 运行文档

- **安装指南**: 包括系统要求（Go 版本、Node 版本）、依赖安装、编译步骤。
- **故障排查**: 常见问题与解决方案（如连接超时、房间未找到等）。

## 14. 开发环境与工具

### A. 系统要求

- **Go**: 版本 ≥ **1.19**。
- **Node.js**: 版本 ≥ **16.14**，建议使用 **Node 18 LTS** 或更高。
- **npm/yarn**: 包管理器，推荐 npm **8.x+**。
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`。

### B. IDE 与扩展

- **VS Code 推荐扩展**:
  - `Go` (golang.go)：Go 语言支持。
  - `Vetur` 或 `Volar`：Vue 3 支持（推荐后者）。
  - `TypeScript Vue Plugin`：Vue + TypeScript 集成。
  - `Tailwind CSS IntelliSense`：Tailwind 样式提示。
- **调试工具**:
  - Go: Delve 调试器（VS Code 集成）。
  - 前端: Chrome DevTools（F12）。

### C. 依赖版本锁定

- **Go**: 在 `go.mod` 中明确指定所有依赖版本，不使用浮点版本号（如 `v2.1.x`）。
- **npm**: 使用 `package-lock.json` 锁定前端依赖版本，运行 `npm ci` 而非 `npm install` 进行安装。

## 15. 调试与监控

### A. 本地调试

- **Go 调试**: 在 VS Code 中配置 `.vscode/launch.json` 用于 Wails 调试。
- **前端调试**: 使用 Vue DevTools 检查组件状态，使用 Network 标签页监控 WebSocket 流量。
- **性能分析**: 使用 Chrome DevTools 的 Performance 标签进行帧率分析，使用 Lighthouse 检查 Web 性能。

### B. 日志详细度控制

- **Go**: 可通过环境变量 `LOG_LEVEL` 控制日志级别（DEBUG, INFO, WARN, ERROR）。
- **前端**: 可通过 localStorage `DEBUG_MODE` 标志启用详细日志。

## 15.5 管理员面板系统设计

### A. 管理员身份验证流程

1. **密码设置**:

   - 在配置文件 `config.yml` 中设置管理员密码
   - 生产环境建议使用 bcrypt 等算法进行哈希存储
   - 支持环境变量覆盖：`ADMIN_PASSWORD=yourpassword`

2. **验证机制**:

   - 前端：用户点击"管理员入口"，弹出密码输入框
   - 调用后端 `VerifyAdminPassword(roomId, password)` 方法
   - 后端验证成功后下发短期会话令牌（Session Token，有效期 30 分钟）
   - 前端存储令牌在 sessionStorage 中（页面关闭自动清除）

3. **会话管理**:
   - 每次请求都在 WebSocket 消息头中携带令牌
   - 后端在每次管理员操作前校验令牌有效性
   - 令牌过期时前端自动重新验证

### B. 玩家列表视图设计

**表格列**（从左到右）:

- **序号**: 简单递增序列
- **玩家名称**: 点击可弹出详情
- **玩家 ID**: UUID 的缩写（前 8 位）或复制按钮
- **状态**: 在线 / 闲置 / 已死亡（用不同颜色标记）
- **血量**: 血条进度 + 数值（e.g., 75/100）
- **位置**: {x, z} 坐标或"隐藏"
- **延迟**: Ping 值（绿 < 100ms，黄 100-300ms，红 > 300ms）
- **操作**: 观察玩家、查看详情、踢出等按钮

**功能**:

- 实时刷新（每 250ms 更新一次）
- 支持模糊搜索（按名称、ID 搜索）
- 支持按列排序（默认按昵称）
- 显示总玩家数 / 最大玩家数

### C. 玩家详情弹窗设计

弹窗分为 4 个 Tab：

**Tab 1 - 基础信息**:

- 玩家昵称、UUID、加入时间、在线时长
- 当前位置（X, Y, Z 坐标）
- 当前方向（Pitch, Yaw 欧拉角）
- 移动速度（向量形式）
- 血量 / 最大血量
- 状态：在线 / 闲置 / 已死亡

**Tab 2 - 装备与背包**:

- 当前装备：武器、护甲、弹药
- 背包物品列表（网格形式，每行 5 个）
  - 物品图标、名称、数量
  - 可点击查看物品详情

**Tab 3 - 连接信息 & 统计**:

- **网络**:

  - 远程 IP 地址
  - Ping 值（实时更新）
  - 丢包率 (%)
  - 连接持续时间

- **游戏统计**:
  - 放置方块数
  - 移除方块数
  - 击杀数 / 死亡数
  - 移动距离（估值）

**Tab 4 - 管理操作**（可选）:

- **观察玩家**: 切换摄像头跟随该玩家视角
- **踢出玩家**: 带可选离线原因
- **禁言玩家**: 选择禁言时长（5min / 30min / 1hour / 永久）
- **查看操作日志**: 显示该玩家的最近操作记录

### D. 房间统计面板

显示当前房间的聚合数据：

```
┌─ 房间统计 ────────────────────┐
│ 房间号: 123456                 │
│ 在线玩家: 5 / 10               │
│ 平均延迟: 45ms                 │
│ 房间运行时间: 2h 15m           │
│ 总方块操作: 12,450            │
│  - 放置: 8,230                │
│  - 移除: 4,220                │
│ 峰值玩家数: 8                  │
│ 创建时间: 2024-02-10 14:30     │
└────────────────────────────────┘
```

### E. 前端状态管理 (Pinia Store)

```typescript
// stores/admin.ts
interface AdminState {
  isAuthenticated: boolean;      // 是否已验证
  sessionToken: string | null;   // 会话令牌
  tokenExpiresAt: number;        // 令牌过期时间
  players: PlayerInfo[];         // 玩家列表
  selectedPlayer: PlayerDetails | null; // 当前选中玩家详情
  roomStats: RoomStatistics | null;
  lastUpdateTime: number;
  isLoading: boolean;
  error: string | null;
}

// Actions
verifyPassword(password: string);
logout();
refreshPlayerList();
getPlayerDetails(playerId: string);
kickPlayer(playerId: string, reason?: string);
mutePlayer(playerId: string, durationSeconds: number);
observePlayer(playerId: string);
```

### F. 后端 API 实现要点

**`VerifyAdminPassword` 实现**:

```python-pseudocode
1. 接收房间ID和明文密码
2. 从配置加载密码哈希值
3. 使用 bcrypt.Compare() 验证
4. 验证成功:
   - 生成 JWT Token（有效期 30 分钟）
   - 返回 Token 和过期时间戳
5. 验证失败:
   - 记录尝试日志（防暴力破解）
   - 返回错误信息
```

**`GetOnlinePlayers` 实现**:

```python-pseudocode
1. 校验 SessionToken 有效性
2. 遍历房间内所有连接客户端
3. 收集各客户端的玩家信息（PlayerInfo）
4. 计算 Ping 值（基于心跳包往返时间）
5. 按最后更新时间降序排列
6. 返回列表
```

**`GetPlayerDetails` 实现**:

```python-pseudocode
1. 校验 SessionToken 有效性
2. 根据 playerId 查找玩家
3. 收集完整的 PlayerDetails 数据
4. 包括装备、背包、连接统计、游戏统计等
5. 返回序列化后的详情对象
```

### G. 前端组件通信流程

```
AdminPanel.vue
  ├─ AdminLogin.vue (密码验证)
  │  └─ 调用 AdminService.verifyPassword()
  │     └─ 派发 admin store action: verifyPassword
  │
  ├─ PlayerListView.vue (玩家列表)
  │  ├─ 定时调用 AdminService.getOnlinePlayers()
  │  ├─ 更新 admin store: players
  │  └─ 点击玩家行 → 显示 PlayerDetailModal
  │
  └─ PlayerDetailModal.vue (详情弹窗)
      ├─ 接收 selectedPlayer 从 admin store
      ├─ 显示 4 个 Tab
      └─ 管理员操作按钮调用相应 API
         ├─ kickPlayer()
         ├─ mutePlayer()
         └─ observePlayer()
```

## 16. 详细数据格式与通信协议

### A. WebSocket 消息 Envelope 规范

所有 WebSocket 消息采用统一 Envelope 格式（JSON）：

```json
{
  "type": "string", // 消息类型，见下方清单
  "timestamp": 1234567890, // 毫秒级时间戳
  "playerId": "uuid-string", // 发送者玩家 ID
  "roomId": "123456", // 房间号
  "id": "msg-uuid", // 消息唯一标识（用于去重、确认）
  "payload": {} // 具体数据，根据 type 而定
}
```

### B. 核心消息类型详解

| 消息类型 | 方向 | 说明 | 示例 payload |
| --- | --- | --- | --- |
| `player_join` | 服务→全客 | 玩家加入房间 | `{playerId, playerName, position: {x,y,z}}` |
| `player_leave` | 服务→全客 | 玩家离开房间 | `{playerId}` |
| `player_move` | 客→服/服→全客 | 玩家位置或旋转更新 | `{x, y, z, rotation: {pitch, yaw}}` |
| `player_state_sync` | 服→客 | 同步其他玩家状态（插值参数） | `{players: [{id, x, y, z, vx, vy, vz, rotation}]}` |
| `block_place` | 客→服 | 放置方块请求 | `{x, y, z, blockType: "stone\|wood\|glass"}` |
| `block_remove` | 客→服 | 移除方块请求 | `{x, y, z}` |
| `world_update` | 服→客 | 世界变化广播 | `{changes: [{x, y, z, blockType, action: "place\|remove"}]}` |
| `world_snapshot` | 服→客 | 完整世界状态（初始化） | `{chunk: {x, z}, blocks: [...]}` |
| `ping` | 双向 | 心跳检测 | `{}` |
| `pong` | 双向 | 心跳响应 | `{}` |

### C. 数据类型定义

**玩家状态 (PlayerState):**

```typescript
interface PlayerState {
  id: string; // UUID
  name: string; // 玩家昵称
  position: Vector3; // 当前位置 {x, y, z}
  velocity: Vector3; // 速度向量（用于插值）
  rotation: Rotation; // 欧拉角 {pitch, yaw, roll}
  health: number; // 血量 0-100
  ammo: number; // 弹药数量
  equipment: string; // 装备类型 ("pistol"|"rifle"|"shotgun")
  isAlive: boolean; // 是否存活
  lastUpdateTime: number; // 最后更新时间戳
}
```

**方块数据 (BlockData):**

```typescript
interface BlockData {
  x: number;
  y: number;
  z: number; // 体素坐标
  type: string; // 方块类型 ("stone"|"wood"|"glass"|"dirt")
  metadata?: number; // 元数据（旋转、颜色等，可选）
}
```

**房间配置 (RoomConfig):**

```typescript
interface RoomConfig {
  roomId: string; // 6位数字房间号
  port: number; // WebSocket 监听端口
  maxPlayers: number; // 最大玩家数
  currentPlayers: number; // 当前玩家数
  worldSeed: string; // 世界种子（用于生成）
  createdAt: number; // 创建时间戳
  lastActivityAt: number; // 最后活动时间戳
  isPublic: boolean; // 是否公开（LAN 发现）
  gameMode: string; // 游戏模式 ("sandbox"|"survival"|"pvp")
}
```

**玩家列表信息 (PlayerInfo - 用于管理员列表):**

```typescript
interface PlayerInfo {
  id: string; // 玩家 UUID
  name: string; // 玩家昵称
  position: Vector3; // 当前位置
  health: number; // 血量 (0-100)
  status: 'online' | 'idle' | 'dead'; // 玩家状态
  connectedAt: number; // 连接时间戳
  lastActivityAt: number; // 最后活动时间戳
  ping: number; // 网络延迟 (ms)
  equipment: string; // 当前装备
}
```

**玩家详细信息 (PlayerDetails - 用于详情窗口):**

```typescript
interface PlayerDetails {
  // 基础信息
  id: string;
  name: string;

  // 位置和状态
  position: Vector3;
  velocity: Vector3;
  rotation: Rotation;
  health: number;
  maxHealth: number;
  isAlive: boolean;
  status: 'online' | 'idle' | 'dead';

  // 装备与背包
  equipment: {
    weapon: string; // 武器类型
    armor: string; // 护甲类型
    ammo: number; // 弹药数量
  };
  inventory: Array<{
    // 背包物品
    itemId: string;
    itemType: string;
    quantity: number;
    metadata?: any;
  }>;

  // 连接信息
  connectedAt: number;
  lastActivityAt: number;
  remoteIP: string; // 玩家 IP 地址
  ping: number; // 网络延迟 (ms)
  packetLoss: number; // 丢包率 (%)

  // 统计信息
  statistics: {
    blocksPlaced: number; // 放置方块数
    blocksRemoved: number; // 移除方块数
    killCount: number; // 击杀数量
    deathCount: number; // 死亡次数
    distanceTraveled: number; // 移动距离
  };

  // 管理信息
  isMuted: boolean; // 是否被禁言
  muteEndTime?: number; // 禁言结束时间戳
  joinedAt: number; // 加入房间时间戳
}
```

**房间统计数据 (RoomStatistics):**

```typescript
interface RoomStatistics {
  roomId: string;
  totalPlayers: number; // 当前在线玩家数
  maxPlayers: number; // 最大玩家数
  totalPlayersJoined: number; // 总共加入过的玩家数
  uptime: number; // 房间存在时长 (秒)
  totalBlocksPlaced: number; // 房间内放置方块总数
  totalBlocksRemoved: number; // 房间内移除方块总数
  averagePing: number; // 平均延迟
  peakPlayerCount: number; // 峰值玩家数
  createdAt: number;
  lastUpdated: number;
}
```

**3D 模型信息 (ModelInfo):**

```typescript
interface ModelInfo {
  modelId: string; // 模型唯一标识符
  name: string; // 模型名称
  format: string; // 文件格式 (GLTF, GLB, FBX, OBJ, DAE)
  fileSize: number; // 文件大小 (字节)
  filePath: string; // 本地存储路径
  md5Hash: string; // 文件 MD5 哈希值
  uploadedAt: number; // 上传时间戳
  uploadedBy: string; // 上传者 ID
  thumbnailUrl?: string; // 缩略图 URL（可选）
  metadata?: {
    vertexCount?: number; // 顶点数
    triangleCount?: number; // 三角形数
    materials?: number; // 材质数量
    textures?: number; // 纹理数量
    hasAnimations?: boolean; // 是否包含动画
  };
  version: number; // 模型版本号
  isPublic: boolean; // 是否在局域网共享
}
```

**玩家资料 (PlayerProfile):**

```typescript
interface PlayerProfile {
  playerId: string; // 玩家 UUID
  nickname: string; // 昵称
  avatar?: string; // 头像 URL（可选）
  characterModel: string; // 当前角色模型 GLB ID
  joinedAt: number; // 加入游戏时间戳
  lastSeenAt: number; // 最后在线时间戳
  totalPlayTime: number; // 总游玩时长（秒）
  level: number; // 玩家等级
  experience: number; // 经验值
  nextLevelExp: number; // 下一级所需经验

  // 装扮与定制
  customization: {
    skinColor?: string; // 皮肤颜色 (HEX)
    clothingStyle?: string; // 衣着风格 ID
    accessories?: string[]; // 装饰品 ID 数组
  };

  // 装备与背包
  equipment: {
    weapon?: string; // 当前武器模型 ID
    armor?: string; // 护甲模型 ID
    ammo: number; // 弹药数量
  };
  inventory: Array<{
    // 背包物品
    itemId: string;
    itemType: string;
    quantity: number;
    metadata?: any;
  }>;
}
```

**玩家统计 (PlayerStatistics):**

```typescript
interface PlayerStatistics {
  playerId: string;
  totalBlocksPlaced: number; // 放置方块总数
  totalBlocksRemoved: number; // 移除方块总数
  totalKills: number; // 总击杀数
  totalDeaths: number; // 总死亡次数
  distanceTraveled: number; // 移动总距离
  gameTime: number; // 游戏时间（秒）
  roomsVisited: number; // 访问的房间数量
  achievements: string[]; // 成就 ID 列表
  lastUpdated: number;
}
```

## 17. 项目文件组织结构（详细版）

```
CraftFire/
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── GameScene.vue
│   │   │   ├── GameHUD.vue
│   │   │   ├── RoomCreate.vue
│   │   │   ├── RoomJoin.vue
│   │   │   ├── LANServerList.vue
│   │   │   ├── Player/
│   │   │   │   ├── PlayerEntity.vue
│   │   │   │   ├── FirstPersonController.vue
│   │   │   │   └── RemotePlayer.vue
│   │   │   ├── World/
│   │   │   │   ├── VoxelWorld.vue
│   │   │   │   ├── BlockPlacer.vue
│   │   │   │   ├── Raycaster.vue
│   │   │   │   └── ChunkManager.vue
│   │   │   ├── UI/
│   │   │   │   ├── MainMenu.vue
│   │   │   │   ├── PauseMenu.vue
│   │   │   │   ├── SettingsPanel.vue
│   │   │   │   ├── Toast.vue
│   │   │   │   ├── LoadingBar.vue
│   │   │   │   ├── PlayerProfile.vue         # 个人页面主容器
│   │   │   │   ├── ModelImporter.vue        # 3D 模型导入工具
│   │   │   │   ├── ModelViewer.vue          # 3D 模型查看器
│   │   │   │   ├── Admin/
│   │   │   │   │   ├── AdminPanel.vue      # 管理员面板主容器
│   │   │   │   │   ├── AdminLogin.vue     # 管理员密码验证
│   │   │   │   │   ├── PlayerListView.vue # 玩家列表视图（表格）
│   │   │   │   │   ├── PlayerDetailModal.vue # 玩家详情弹窗
│   │   │   │   │   └── RoomStatsPanel.vue # 房间统计面板
│   │   │   ├── Common/
│   │   │   │   ├── Button.vue
│   │   │   │   ├── Modal.vue
│   │   │   │   └── Input.vue
│   │   │   └── Effects/
│   │   │       ├── ParticleEffect.vue
│   │   │       └── BloodEffect.vue
│   │   ├── stores/
│   │   │   ├── gameState.ts
│   │   │   ├── websocket.ts
│   │   │   ├── player.ts
│   │   │   ├── room.ts
│   │   │   ├── world.ts
│   │   │   ├── ui.ts
│   │   │   ├── settings.ts
│   │   │   ├── admin.ts              # 管理员面板状态管理
│   │   │   ├── profile.ts            # 个人页面状态管理
│   │   │   └── model.ts              # 3D 模型导入状态管理
│   │   ├── services/
│   │   │   ├── WailsService.ts
│   │   │   ├── WebSocketService.ts
│   │   │   ├── InterpolationEngine.ts
│   │   │   ├── InputHandler.ts
│   │   │   ├── PhysicsService.ts
│   │   │   ├── AudioService.ts
│   │   │   ├── AdminService.ts         # 管理员相关API调用
│   │   │   ├── ModelService.ts         # 3D 模型导入、下载、管理
│   │   │   └── ProfileService.ts       # 个人信息获取与更新
│   │   ├── types/
│   │   │   ├── game.ts
│   │   │   ├── player.ts
│   │   │   ├── world.ts
│   │   │   ├── websocket.ts
│   │   │   ├── room.ts
│   │   │   ├── admin.ts               # 管理员相关类型定义
│   │   │   ├── model.ts               # 3D 模型相关类型定义
│   │   │   ├── profile.ts             # 个人信息类型定义
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   ├── math.ts
│   │   │   ├── vector.ts
│   │   │   ├── voxel.ts
│   │   │   ├── uuid.ts
│   │   │   ├── time.ts
│   │   │   └── logger.ts
│   │   ├── composables/
│   │   │   ├── useCamera.ts
│   │   │   ├── useInput.ts
│   │   │   ├── useWebSocket.ts
│   │   │   ├── useGame.ts
│   │   │   ├── useModelViewer.ts        # 3D 模型查看器（旋转、缩放、平移）
│   │   │   └── useProfileData.ts        # 个人页面数据钩子
│   │   ├── assets/
│   │   │   ├── models/
│   │   │   │   ├── player.glb
│   │   │   │   ├── weapons/
│   │   │   │   └── environments/
│   │   │   ├── textures/
│   │   │   │   ├── blocks/
│   │   │   │   └── ui/
│   │   │   ├── sounds/
│   │   │   │   ├── footsteps.mp3
│   │   │   │   ├── fire.mp3
│   │   │   │   └── ambient.mp3
│   │   │   └── styles/
│   │   │       └── variables.css
│   │   ├── App.vue
│   │   └── main.ts
│   ├── public/
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   └── package.json
│
├── backend/
│   ├── main.go
│   ├── app.go
│   ├── cmd/
│   │   └── version.go                 // 版本信息
│   ├── internal/
│   │   ├── websocket/
│   │   │   ├── hub.go
│   │   │   ├── client.go
│   │   │   ├── server.go
│   │   │   ├── message.go
│   │   │   └── handlers.go
│   │   ├── room/
│   │   │   ├── manager.go
│   │   │   ├── room.go
│   │   │   └── config.go
│   │   ├── player/
│   │   │   ├── state.go
│   │   │   ├── manager.go
│   │   │   └── interpolation.go
│   │   ├── world/
│   │   │   ├── voxel.go
│   │   │   ├── chunk.go
│   │   │   ├── generator.go
│   │   │   ├── physics.go
│   │   │   └── persistence.go
│   │   ├── lan/
│   │   │   ├── discovery.go
│   │   │   └── broadcaster.go
│   │   ├── model/
│   │   │   ├── manager.go                  # 3D 模型管理（导入、存储、查询）
│   │   │   ├── importer.go                 # 模型导入处理（格式验证、转换）
│   │   │   ├── storage.go                  # 本地模型存储与持久化
│   │   │   ├── validator.go                # 模型文件校验（大小、格式、MD5）
│   │   │   └── sync.go                     # 局域网模型同步（断点续传、版本管理）
│   │   ├── admin/
│   │   │   ├── authentication.go       # 管理员密码验证与授权
│   │   │   └── stats.go                # 房间统计数据收集
│   │   ├── profile/
│   │   │   ├── provider.go             # 玩家个人信息提供者
│   │   │   └── characterRenderer.go    # 角色模型渲染数据组装
│   │   ├── config/
│   │   │   ├── config.go
│   │   │   └── defaults.go
│   │   ├── logger/
│   │   │   └── logger.go
│   │   └── utils/
│   │       ├── uuid.go
│   │       └── math.go
│   ├── tests/
│   │   ├── websocket_test.go
│   │   ├── room_test.go
│   │   ├── player_test.go
│   │   └── world_test.go
│   ├── go.mod
│   └── go.sum
│
├── .gitignore
├── .vscode/
│   ├── launch.json                    // 调试配置
│   ├── settings.json                  // 开发设置
│   └── extensions.json                // 推荐扩展
├── docs/
│   ├── README.md                      // 项目介绍
│   ├── ARCHITECTURE.md                // 架构设计
│   ├── CONTRIBUTING.md                // 贡献指南
│   ├── API.md                         // WebSocket API 文档
│   ├── INSTALL.md                     // 安装指南
│   └── TROUBLESHOOTING.md             // 故障排查
├── config.example.yml                 // 配置示例
├── Makefile                           // 便利脚本
├── docker-compose.yml                 // Docker 配置（可选）
└── wails.json                         // Wails 配置
```

## 18. 环境配置与启动命令

### A. 快速开发启动

```bash
# 初始化项目
git clone <repo-url>
cd CraftFire

# 安装依赖
cd frontend && npm install
cd ../backend && go mod tidy

# 启动开发服务（Wails 自动处理前后端）
wails dev

# 或分离启动
# 终端 1 - 前端
cd frontend && npm run dev

# 终端 2 - 后端
cd backend && wails dev
```

### B. 常用构建命令

```bash
# 构建可执行文件
wails build --output CraftFire.exe

# 生成 Windows MSI 安装程序
wails build -nsis

# 优化构建（删除调试符号，压缩）
wails build -ldflags=-s -ldflags=-w

# 清理构建缓存
wails build -clean
```

### C. 测试命令

```bash
# Go 单元测试
cd backend && go test ./...

# Go 测试覆盖率
go test -cover ./...

# 前端单元测试
cd frontend && npm run test:unit

# 前端 E2E 测试
npm run test:e2e

# 性能基准测试（Go）
go test -bench ./... -benchmem
```

## 19. 命名规范（详细版）

### A. Go 命名

- **文件名**: `snake_case.go`（例如 `websocket_hub.go`, `player_manager.go`）
- **常量**: `SCREAMING_SNAKE_CASE`（例如 `MAX_PLAYERS = 50`, `DEFAULT_TICK_RATE = 8`）
- **函数/方法**: `PascalCase`（例如 `CreateRoom()`, `ProcessMessage()`）
- **变量**: `camelCase`（例如 `playerState`, `roomManager`）
- **包名**: 全小写简洁（例如 `websocket`, `room`, `player`）
- **接口**: `PascalCase` 且通常以 `r` 或 `er` 结尾（例如 `Reader`, `Manager`）

### B. TypeScript/Vue 命名

- **文件名**: PascalCase（例如 `GameScene.vue`, `PlayerEntity.vue`）或 camelCase（例如 `websocket.ts`, `gameState.ts`）
- **组件**: PascalCase（始终使用 PascalCase，如 `<GameHUD />`, `<FirstPersonController />`）
- **变量/函数**: camelCase（例如 `playerPosition`, `calculateDistance()`）
- **常量**: SCREAMING_SNAKE_CASE（例如 `MAX_BLOCK_STACK = 64`）
- **接口**: PascalCase 或 `I<Name>` 前缀（例如 `PlayerState` 或 `IPlayerState`）
- **类型别名**: PascalCase（例如 `type Vector3 = {x, y, z}`）

### C. 类/结构体成员

- **Go struct 字段**: PascalCase（导出）或 camelCase（私有，Go 中需要首字母小写）
- **TypeScript class 成员**: private 前缀 `_` 加 camelCase（例如 `_health`, `_position`）

## 20. 代码注释规范（中文）

### A. JSDoc 注释示例

```typescript
/**
 * 计算两个三维点之间的欧几里得距离
 *
 * @param p1 - 第一个点的坐标 {x, y, z}
 * @param p2 - 第二个点的坐标 {x, y, z}
 * @param useSquareRoot - 是否计算平方根，默认 true；如为 false 返回平方距离（更快）
 * @returns 距离值
 *
 * @example
 * const dist = distance3D({x: 0, y: 0, z: 0}, {x: 3, y: 4, z: 0});
 * console.log(dist); // 5
 */
export function distance3D(
  p1: Vector3,
  p2: Vector3,
  useSquareRoot = true,
): number {
  // 实现注释
  const dx = p2.x - p1.x;
  const dy = p2.y - p1.y;
  const dz = p2.z - p1.z;
  const sqDist = dx * dx + dy * dy + dz * dz;
  return useSquareRoot ? Math.sqrt(sqDist) : sqDist;
}
```

### B. GoDoc 注释示例

```go
// CalculateDistance 计算两个三维点之间的欧几里得距离。
//
// 参数：
// - p1, p2: 三维坐标点
// - useSquareRoot: 是否计算平方根，false 时返回平方距离以提高性能
//
// 返回值：距离值（float32）
//
// 示例：
//  dist := CalculateDistance(Vec3{0, 0, 0}, Vec3{3, 4, 0}, true)
//  fmt.Println(dist) // 5.0
func CalculateDistance(p1, p2 Vec3, useSquareRoot bool) float32 {
  dx := p2.X - p1.X
  dy := p2.Y - p1.Y
  dz := p2.Z - p1.Z
  sqDist := dx*dx + dy*dy + dz*dz
  if !useSquareRoot {
    return sqDist
  }
  return float32(math.Sqrt(float64(sqDist)))
}
```

### C. 行内注释规则

- 仅在逻辑复杂、不显而易见的地方添加行内注释
- 注释应解释"为什么"，而非"做什么"
- 格式: `// 原因说明` 或 /_ 多行原因 _/

```typescript
// 不好
const x = 10; // 将 x 赋值为 10

// 好
const maxRetries = 5; // 根据 Redis 超时时间（1 秒），允许最多 5 次重试

// 不好
if (player.health <= 0) {
  // 检查玩家是否死亡
  player.remove();
}

// 好
if (player.health <= 0) {
  player.remove(); // 尸体在下一帧由垃圾回收处理
}
```

## 21. 配置管理（Config Management）

### A. 配置文件 (`config.yml`)

```yaml
# ==========================================
# CraftFire 应用配置
# ==========================================

# 应用信息
app:
  name: 'CraftFire'
  version: '1.0.0'
  description: 'Sandbox + FPS Hybrid Game'

# 服务器配置
server:
  defaultTickRate: 8 # 刷新率 (ticks/sec)
  maxRoomsPerIP: 5 # 单 IP 最多创建房间数
  maxConcurrentPlayers: 200 # 全局最大玩家数
  roomIdleTimeout: 3600 # 房间空置 1 小时后自动关闭 (秒)
  playerIdleTimeout: 300 # 玩家 5 分钟未响应自动断开 (秒)

# LAN 发现配置
lan:
  broadcastInterval: 5 # UDP 广播间隔 (秒)
  discoveryPort: 9999 # UDP 发现端口
  broadcastAddr: '255.255.255.255' # 广播地址

# 世界配置
world:
  chunkSize: 16 # 分块大小 (体素单位)
  renderDistance: 8 # 渲染距离 (分块数)
  worldHeightLimit: 256 # 世界高度 (块)
  persistenceInterval: 300 # 持久化间隔 (秒)
  autoSave: true

# 物理引擎配置
physics:
  gravity: 9.8 # 重力加速度
  dragCoefficient: 0.1 # 空气阻力系数
  maxVelocity: 100 # 最大速度限制

# 前端配置
client:
  targetFPS: 60 # 目标帧率
  interpolationDuration: 125 # 插值持续时间 (ms)，应与刷新率对应
  pointerLockRequired: true # 是否要求指针锁定
  defaultRenderDistance: 6 # 默认渲染距离

# UI 配置
ui:
  language: 'zh-CN' # 语言
  uiScale: 1.0 # UI 缩放
  showDebugInfo: false # 是否显示调试信息

# 日志配置
logging:
  level: 'INFO' # DEBUG, INFO, WARN, ERROR
  format: 'text' # text 或 json
  output: 'console' # console 或 file
  filePath: './logs/craftfire.log'
  maxLogSize: 10485760 # 单个日志文件最大大小 (10MB)
  maxLogBackups: 5 # 保留日志文件数

# 安全配置
security:
  roomPasswordRequired: false
  enableTLS: false # 仅局域网，通常 false
  rateLimitPerSecond: 100 # 每秒消息速率限制
  adminPassword: 'admin123' # 管理员面板密码（生产环境应使用强密码）
  adminPasswordHash: '' # 密码哈希值（使用 bcrypt 等安全算法）
  enableAdminPanel: true # 是否启用管理员面板
```

### B. 环境变量

```bash
# 开发环境 (.env.development)
VITE_API_BASE_URL=ws://localhost:8080
VITE_DEBUG_MODE=true
VITE_LOG_LEVEL=DEBUG

# 生产环境 (.env.production)
VITE_API_BASE_URL=ws://127.0.0.1:8080
VITE_DEBUG_MODE=false
VITE_LOG_LEVEL=WARN
```

## 22. 模组与扩展系统（可选但推荐）

### A. 模组加载机制

- **模组目录**: `~/.CraftFire/mods/` 或项目内 `resources/mods/`
- **模组格式**: 目录形式，每个模组为单独文件夹
- **模组入口**: `mod.json` 清单文件 + Lua 脚本 + 资源

```json
{
  "name": "custom_blocks",
  "version": "1.0.0",
  "author": "user",
  "description": "自定义方块类型模组",
  "entryPoint": "main.lua",
  "dependencies": [],
  "minGameVersion": "1.0.0"
}
```

### B. 支持的拓展类型

- **方块类型**: 自定义方块属性（纹理、硬度、掉落物等）
- **工具/武器**: 新增游戏内工具和武器
- **命令**: 自定义游戏内指令
- **事件钩子**: 响应游戏事件（玩家加入、建造、死亡等）
- **UI 组件**: 自定义 HUD 元素

### C. 热加载

- 支持无需重启应用即可加载/卸载模组
- 使用观察器监控 `mods/` 目录变化
- 类似 Minecraft 模组管理系统体验

## 23. 任务清单（优先级）

请生成项目结构以及以下核心代码实现，按优先级排列：

**P0 - 必须实现（MVP）:**

1. **Go** `main.go`: 应用入口，Wails 应用初始化
2. **Go** `internal/websocket/hub.go`: WebSocket Hub，核心消息广播
3. **Go** `internal/room/manager.go`: 房间管理（创建、加入、销毁）
4. **Frontend** `src/components/GameScene.vue`: 主 3D 场景（TresJS）
5. **Frontend** `src/stores/websocket.ts`: WebSocket 连接与状态管理（Pinia）
6. **Frontend** `src/components/RoomCreate.vue` + `RoomJoin.vue`: 房间界面

**P1 - 高优先级（核心功能）:**

7. **Go** `internal/websocket/server.go`: WebSocket 服务启动
8. **Go** `internal/player/manager.go`: 玩家状态管理
9. **Frontend** `src/components/Player/FirstPersonController.vue`: 第一人称控制
10. **Frontend** `src/services/InterpolationEngine.ts`: 插值平滑算法
11. **Frontend** `src/services/WebSocketService.ts`: WebSocket 客户端封装

**P2 - 中优先级（完整体验）:**

12. **Go** `internal/lan/discovery.go`: LAN 发现机制
13. **Frontend** `src/components/Player/RemotePlayer.vue`: 远端玩家渲染
14. **Frontend** `src/services/InputHandler.ts`: 键盘鼠标输入处理
15. **Frontend** `src/types/game.ts`: TypeScript 类型完整定义

**P2.5 - 管理员面板（新需求）:**

16. **Frontend** `src/components/UI/Admin/AdminPanel.vue`: 管理员面板主容器
17. **Frontend** `src/components/UI/Admin/AdminLogin.vue`: 管理员密码验证
18. **Frontend** `src/components/UI/Admin/PlayerListView.vue`: 在线玩家列表（表格）
19. **Frontend** `src/components/UI/Admin/PlayerDetailModal.vue`: 玩家详情弹窗
20. **Frontend** `src/components/UI/Admin/RoomStatsPanel.vue`: 房间统计面板
21. **Frontend** `src/stores/admin.ts`: 管理员状态管理（Pinia）
22. **Frontend** `src/services/AdminService.ts`: 管理员 API 调用封装
23. **Frontend** `src/types/admin.ts`: 管理员相关 TypeScript 类型
24. **Go** `internal/admin/authentication.go`: 管理员密码验证与授权
25. **Go** `internal/admin/stats.go`: 房间统计数据收集与计算

**P2.75 - 3D 文件导入与个人页面（新需求）:**

26. **Frontend** `src/components/UI/ModelTools/ModelImporter.vue`: 3D 模型导入工具
27. **Frontend** `src/components/UI/ModelTools/ModelViewer.vue`: 3D 模型查看器
28. **Frontend** `src/components/UI/PlayerProfile.vue`: 个人页面主容器
29. **Frontend** `src/stores/model.ts`: 3D 模型状态管理
30. **Frontend** `src/stores/profile.ts`: 个人页面状态管理
31. **Frontend** `src/services/ModelService.ts`: 模型导入、上传、下载、管理
32. **Frontend** `src/services/ProfileService.ts`: 个人信息获取与更新
33. **Frontend** `src/composables/useModelViewer.ts`: 3D 模型查看器交互逻辑
34. **Frontend** `src/composables/useProfileData.ts`: 个人页面数据获取钩子
35. **Frontend** `src/types/model.ts`: 3D 模型相关 TypeScript 类型
36. **Frontend** `src/types/profile.ts`: 个人信息相关 TypeScript 类型
37. **Go** `internal/model/manager.go`: 3D 模型管理系统
38. **Go** `internal/model/importer.go`: 模型导入处理
39. **Go** `internal/model/storage.go`: 本地模型存储
40. **Go** `internal/model/validator.go`: 模型文件校验
41. **Go** `internal/model/sync.go`: 局域网模型同步

**P3 - 低优先级（优化与扩展）:**

42. **Go** `internal/world/voxel.go`: 体素数据结构
43. **Frontend** `src/components/World/VoxelWorld.vue`: 体素世界渲染
44. **Frontend** `src/components/GameHUD.vue`: 游戏内 HUD
45. 项目文档（README.md, ARCHITECTURE.md 等）
46. **Go** `internal/admin/moderation.go`: 踢除玩家、禁言等管理操作（可选扩展）

---

24. 计算分配（Go vs JS）与工程改动建议 🔬

- 目标：在保证一致性与反作弊能力的前提下，兼顾客户端低延迟渲染体验与服务器的高并发确定性计算能力。

- 核心决策（推荐）：

  - 采用 **Server‑authoritative（后端权威）** + **客户端预测/插值** 的混合模型（后端负责规则与关键物理校验，前端负责高频渲染与预测）。
  - 对于 CPU 密集但可并行的离线/预处理任务（如体素世界生成、LOD 计算、网格烘焙），优先使用 **WASM**（放前端或作为可选工具）；对实时且必须一致的逻辑（伤害计算、方块放置最终判定、世界快照合并）放 **Go 后端**。

- 需要在提示词中补充的工程项（便于开发者快速落地）：

  - 后端（必需）
    - 新增包：`internal/physics/engine.go`（权威物理/规则校验）
    - 新增文件：`internal/sim/loop.go`（仿真主循环，tick-rate 明确为 8 TPS）
    - 基准与监控：`internal/physics/engine_test.go`（benchmarks），并在 README 中给出 `go test -bench ./... -benchmem` 示例
  - 前端（必需）
    - Worker：`src/workers/physics.worker.ts`（非阻塞的视觉物理 / 碰撞预处理）
    - WASM 桥：`packages/compute-wasm/`（可复用的 C/Rust → wasm 工具和构建脚本）
    - 服务：`src/services/physicsWasm.ts`（WASM 加载与运行封装）
  - 共享工具
    - `packages/utils/voxel-helpers.ts`（体素相关算法）——供前后端复用（Node/Go 需各自实现兼容接口）
  - 配置
    - 在 `config.example.yml` 添加：
      - `physics:`
        - `authority: "server" | "hybrid"`
        - `useWasm: true|false`
        - `workerThreshold: 1000 # 对象数超过该值时启用 worker`
  - CI / 验证
    - 在 CI 增加基准 job：`go test -bench` 与前端的 vitest 性能检查（关键路径帧时间断言）。

- 小的实现/文档片段（可直接复制到提示词或 README）：

  - WebSocket 与校验示例（原则性说明）:

    - 客户端发送：`player_intent { seq, input, predictedPosition }`
    - 服务器响应：`authoritative_state { seqAck, correctedPosition?, corrections[] }`
    - 客户端行为：应用差分并做平滑回滚（interpolation / reconciliation）

  - 推荐新增任务（短清单，可加入到 PR checklist）：
    1. 在后端实现仿真骨架与基准（`internal/physics`）。
    2. 在前端实现 Worker 示例并将现有插值逻辑迁移到 Worker。
    3. 把大量运算（世界生成、网格合并）迁移到 WASM，并在 `packages/` 提供构建脚本。
    4. 在 CI 添加性能回归阈值并编写压力测试脚本（单房间 10–50 用户场景）。

- 验证指标（PR 必填）:
  - 后端：单房间 8 TPS 下 95% 请求处理时间 < 30ms
  - 前端：主线程帧时间 95% < 16ms（在标准硬件）
  - 网络：LAN 往返 < 100ms 常态，峰值 < 200ms

> 小结：把“规则/关键物理”放到 Go（保证一致性）；把“视觉物理/预测/插值/重计算”放到前端（保持 60 FPS）。同时用 Worker/WASM 缓解主线程与提升可移植性能。✅

如果需要，我可以基于上述补充为你：

- 生成后端 `internal/physics` 与基准测试骨架（Go）；或
- 生成前端 `src/workers/physics.worker.ts` + `src/services/physicsWasm.ts` 的示例实现（TS + 简单插值示例）。

请选择要我先生成的代码骨架（后端 / 前端 / 只更新文档）。
