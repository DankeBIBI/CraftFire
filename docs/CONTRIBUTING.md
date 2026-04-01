# CraftFire 贡献指南

感谢你对 CraftFire 的贡献兴趣！请阅读以下指南以保持代码风格统一与项目质量。

## 开发环境

参见 [INSTALL.md](./INSTALL.md) 完成环境搭建。

## 分支策略

| 分支        | 用途         |
| ----------- | ------------ |
| `main`      | 稳定发布分支 |
| `develop`   | 开发集成分支 |
| `feature/*` | 新功能分支   |
| `fix/*`     | 问题修复分支 |
| `docs/*`    | 文档更新分支 |

```
# 创建功能分支
git checkout develop
git pull
git checkout -b feature/my-feature
```

## 代码风格

### Go 后端

- **格式化**: 使用 `gofmt` / `goimports`，保存时自动格式化
- **命名**:
  - 包名: 全小写，单词无分隔 (`websocket`, `player`)
  - 导出函数/类型: `PascalCase` (`CreateRoom`, `PlayerState`)
  - 私有函数/变量: `camelCase` (`handleMessage`, `tickRate`)
  - 常量: `PascalCase` 或 `UPPER_SNAKE_CASE` (`MaxPlayers`, `DEFAULT_TICK_RATE`)
  - 接口名后缀 `er` 或语义化命名 (`Manager`, `Provider`, `Handler`)
- **注释**:
  - 所有导出的函数/类型/常量必须有 GoDoc 注释
  - 行内注释解释「为什么」，而非「做什么」
- **错误处理**: 使用 `fmt.Errorf("context: %w", err)` 包装错误，避免裸 `panic`
- **测试**: 测试文件与源文件同包，命名 `*_test.go`

```go
// ✅ 好
func (m *RoomManager) CreateRoom(config RoomConfig) (*Room, error) {
    // 使用房间号作为端口，仅限局域网场景安全
    port := config.RoomId
    ...
}

// ❌ 差
func (m *RoomManager) CreateRoom(config RoomConfig) (*Room, error) {
    port := config.RoomId // 设置端口
    ...
}
```

### Frontend (Vue 3 + TypeScript)

- **组件**: 使用 `<script setup lang="ts">` + Composition API
- **命名**:
  - 组件文件: `PascalCase.vue` (`GameScene.vue`, `PlayerEntity.vue`)
  - 组合式函数: `use` 前缀 (`useCamera`, `useInput`)
  - Store 文件: `camelCase.ts` (`gameState.ts`, `websocket.ts`)
  - 类型文件: `camelCase.ts`，类型名 `PascalCase`
  - 服务文件: `PascalCase.ts` (`WebSocketService.ts`)
  - 工具文件: `camelCase.ts` (`math.ts`, `vector.ts`)
  - 常量: `UPPER_SNAKE_CASE` (`CHUNK_SIZE`, `MAX_HEALTH`)
- **Props**: 使用 `defineProps<T>()` 泛型形式
- **Emits**: 使用 `defineEmits<T>()` 泛型形式
- **CSS**: 使用 Tailwind CSS 工具类，避免 scoped CSS（除非必要）
- **状态管理**: Pinia，Store 使用 `defineStore` + Composition API 风格

```vue
<!-- ✅ 好 -->
<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useGameState } from "@/stores/gameState";

interface Props {
	playerId: string;
	showDebug?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	showDebug: false,
});

const emit = defineEmits<{
	(e: "update", id: string): void;
}>();

const gameState = useGameState();
</script>
```

### 通用规则

- TypeScript 严格模式，禁止 `any`（除非有明确注释说明原因）
- 使用 `const` 优先，必要时用 `let`，禁止 `var`
- 字符串优先使用单引号
- 行尾无分号（或统一有分号，保持一致即可）
- 文件末尾保留一个空行

## 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Type 列表

| Type       | 说明                   |
| ---------- | ---------------------- |
| `feat`     | 新功能                 |
| `fix`      | 问题修复               |
| `docs`     | 文档改动               |
| `style`    | 代码格式（不影响逻辑） |
| `refactor` | 重构                   |
| `perf`     | 性能优化               |
| `test`     | 测试相关               |
| `chore`    | 构建/工具链变更        |

### Scope 列表

| Scope      | 说明           |
| ---------- | -------------- |
| `backend`  | Go 后端        |
| `frontend` | Vue 前端       |
| `ws`       | WebSocket 相关 |
| `room`     | 房间逻辑       |
| `player`   | 玩家系统       |
| `world`    | 体素世界       |
| `admin`    | 管理员面板     |
| `model`    | 3D 模型系统    |
| `profile`  | 个人页面       |
| `ui`       | UI 组件        |

示例：

```
feat(admin): 添加玩家踢出功能
fix(ws): 修复断线重连后状态未恢复问题
docs(backend): 更新 WebSocket 协议说明
perf(world): 优化分块加载减少内存分配
```

## PR 流程

1. Fork 或创建功能分支
2. 完成开发与测试
3. 确保 `go test ./...` 和 `npm run type-check` 通过
4. 提交 PR 到 `develop` 分支
5. 至少 1 位成员 Review
6. 合并后删除功能分支

## 性能要求

PR 中涉及性能敏感路径时，需满足以下指标：

| 指标                            | 阈值    |
| ------------------------------- | ------- |
| 后端单房间 8 TPS 处理时间 (P95) | < 30ms  |
| 前端主线程帧时间 (P95)          | < 16ms  |
| LAN 往返延迟 (常态)             | < 100ms |
| LAN 往返延迟 (峰值)             | < 200ms |

## 目录结构约定

参见 [ARCHITECTURE.md](./ARCHITECTURE.md) 中的项目结构说明。新增文件应遵循现有目录组织方式。
