# CraftFire WebSocket API 文档

## 概述

CraftFire 使用 WebSocket 进行客户端与服务端之间的实时通信。服务端以 **8 TPS**（每 125ms 一个 tick）的频率广播权威游戏状态，客户端通过插值机制实现 60 FPS 平滑渲染。

- **端口规则**: 6 位数字房间号即为 WS 端口号（如房间 `123456` → `ws://host:123456`）
- **协议**: JSON over WebSocket

## 消息 Envelope

所有消息采用统一信封格式：

```json
{
	"type": "string",
	"timestamp": 1234567890,
	"playerId": "uuid-string",
	"roomId": "123456",
	"id": "msg-uuid",
	"payload": {}
}
```

| 字段        | 类型     | 说明                         |
| ----------- | -------- | ---------------------------- |
| `type`      | `string` | 消息类型标识                 |
| `timestamp` | `number` | 毫秒级时间戳                 |
| `playerId`  | `string` | 发送者 UUID                  |
| `roomId`    | `string` | 6 位房间号                   |
| `id`        | `string` | 消息唯一标识，用于去重/确认  |
| `payload`   | `object` | 具体数据，由 `type` 决定结构 |

## 消息类型

### 连接与心跳

#### `ping` / `pong`

心跳检测，双向发送。

```json
// ping
{ "type": "ping", "timestamp": 1700000000000, "payload": {} }

// pong
{ "type": "pong", "timestamp": 1700000000001, "payload": {} }
```

### 玩家管理

#### `player_join` (服务端 → 全部客户端)

新玩家加入房间通知。

```json
{
	"type": "player_join",
	"payload": {
		"playerId": "uuid-string",
		"playerName": "Player1",
		"position": { "x": 0, "y": 10, "z": 0 }
	}
}
```

#### `player_leave` (服务端 → 全部客户端)

玩家离开房间通知。

```json
{
	"type": "player_leave",
	"payload": {
		"playerId": "uuid-string"
	}
}
```

#### `player_move` (客户端 → 服务端 / 服务端 → 全部客户端)

玩家位置与朝向更新。

```json
{
	"type": "player_move",
	"payload": {
		"x": 10.5,
		"y": 2.0,
		"z": -3.2,
		"rotation": {
			"pitch": 0.1,
			"yaw": 1.57
		}
	}
}
```

#### `player_state_sync` (服务端 → 客户端)

批量同步所有玩家状态，每个 tick (125ms) 广播一次。客户端用此数据进行插值渲染。

```json
{
	"type": "player_state_sync",
	"payload": {
		"players": [
			{
				"id": "uuid-1",
				"x": 10.5,
				"y": 2.0,
				"z": -3.2,
				"vx": 1.0,
				"vy": 0,
				"vz": -0.5,
				"rotation": { "pitch": 0.1, "yaw": 1.57 }
			}
		]
	}
}
```

### 世界操作

#### `block_place` (客户端 → 服务端)

请求放置方块。服务端验证后广播 `world_update`。

```json
{
	"type": "block_place",
	"payload": {
		"x": 5,
		"y": 10,
		"z": -2,
		"blockType": "stone"
	}
}
```

支持的方块类型：`stone` | `wood` | `glass` | `dirt` | `grass` | `sand` | `brick` | `metal`

#### `block_remove` (客户端 → 服务端)

请求移除方块。

```json
{
	"type": "block_remove",
	"payload": {
		"x": 5,
		"y": 10,
		"z": -2
	}
}
```

#### `world_update` (服务端 → 客户端)

世界变化广播，包含一个或多个方块变更。

```json
{
	"type": "world_update",
	"payload": {
		"changes": [
			{ "x": 5, "y": 10, "z": -2, "blockType": "stone", "action": "place" },
			{ "x": 3, "y": 5, "z": 0, "blockType": "", "action": "remove" }
		]
	}
}
```

#### `world_snapshot` (服务端 → 客户端)

完整世界状态快照，玩家首次加入或重连时发送。

```json
{
	"type": "world_snapshot",
	"payload": {
		"chunk": { "x": 0, "z": 0 },
		"blocks": [
			{ "x": 0, "y": 0, "z": 0, "type": "grass" },
			{ "x": 1, "y": 0, "z": 0, "type": "grass" }
		]
	}
}
```

### 管理员消息

#### `admin_kick` (客户端 → 服务端)

管理员踢出玩家请求。需要有效 `sessionToken`。

```json
{
	"type": "admin_kick",
	"payload": {
		"targetPlayerId": "uuid-string",
		"reason": "违规操作",
		"sessionToken": "admin-session-token"
	}
}
```

#### `admin_mute` (客户端 → 服务端)

管理员禁言玩家请求。

```json
{
	"type": "admin_mute",
	"payload": {
		"targetPlayerId": "uuid-string",
		"duration": 300,
		"sessionToken": "admin-session-token"
	}
}
```

### 3D 模型同步

#### `model_list` (服务端 → 客户端)

广播可用模型清单。

```json
{
	"type": "model_list",
	"payload": {
		"models": [
			{
				"id": "model-uuid",
				"name": "weapon_rifle",
				"format": "glb",
				"size": 2048000,
				"hash": "sha256-hash-string"
			}
		]
	}
}
```

#### `model_request` (客户端 → 服务端)

请求下载缺失模型。

```json
{
	"type": "model_request",
	"payload": {
		"modelId": "model-uuid"
	}
}
```

## 客户端-服务端交互流

### Server-Authoritative 模型

```
客户端                          服务端
  │                               │
  │── player_intent ──────────▶   │  客户端发送输入意图
  │   {seq, input, predicted}     │
  │                               │  服务端计算权威状态
  │   ◀── authoritative_state ──  │  返回校正（如果需要）
  │   {seqAck, corrected}        │
  │                               │
  │   客户端应用差分 + 平滑回滚  │
```

### 加入房间流

```
1. 客户端调用 JoinRoom(roomId, ip)
2. Wails 后端连接 ws://ip:roomId
3. 服务端发送 world_snapshot
4. 服务端广播 player_join 给所有客户端
5. 客户端开始接收 player_state_sync (8 TPS)
6. 客户端通过插值渲染 (60 FPS)
```

## Wails 暴露方法 (Go → Frontend)

以下方法通过 Wails Runtime 绑定，前端可直接调用：

| 方法                            | 参数                     | 返回值                   | 说明                   |
| ------------------------------- | ------------------------ | ------------------------ | ---------------------- |
| `CreateRoom(config)`            | `RoomConfig`             | `RoomInfo, error`        | 创建房间并启动 WS 服务 |
| `JoinRoom(roomId, ip)`          | `string, string`         | `RoomInfo, error`        | 加入指定房间           |
| `LeaveRoom()`                   | -                        | `error`                  | 离开当前房间           |
| `FindLANServers()`              | -                        | `[]ServerInfo, error`    | 搜索局域网房间         |
| `GetRoomInfo()`                 | -                        | `RoomInfo, error`        | 获取当前房间信息       |
| `VerifyAdminPassword(pw)`       | `string`                 | `SessionToken, error`    | 验证管理员密码         |
| `GetOnlinePlayers(token)`       | `string`                 | `[]PlayerSummary, error` | 获取在线玩家列表       |
| `GetPlayerDetails(token, id)`   | `string, string`         | `PlayerDetails, error`   | 获取玩家详情           |
| `KickPlayer(token, id, reason)` | `string, string, string` | `error`                  | 踢出玩家               |
| `MutePlayer(token, id, dur)`    | `string, string, int`    | `error`                  | 禁言玩家               |
| `ImportModel(path)`             | `string`                 | `ModelInfo, error`       | 导入 3D 模型           |
| `GetModelList()`                | -                        | `[]ModelInfo, error`     | 获取模型列表           |
| `GetPlayerProfile()`            | -                        | `PlayerProfile, error`   | 获取个人信息           |
| `UpdatePlayerProfile(p)`        | `ProfileUpdate`          | `error`                  | 更新个人信息           |

## 数据类型

### PlayerState

```typescript
interface PlayerState {
	id: string;
	name: string;
	position: { x: number; y: number; z: number };
	velocity: { x: number; y: number; z: number };
	rotation: { pitch: number; yaw: number; roll: number };
	health: number; // 0-100
	ammo: number;
	equipment: string; // "pistol" | "rifle" | "shotgun"
	isAlive: boolean;
	lastUpdateTime: number;
}
```

### BlockData

```typescript
interface BlockData {
	x: number;
	y: number;
	z: number;
	type: string; // "stone" | "wood" | "glass" | "dirt" | ...
	metadata?: number;
}
```

### RoomConfig

```typescript
interface RoomConfig {
	roomId: string;
	port: number;
	maxPlayers: number;
	gameMode: string; // "sandbox" | "survival" | "pvp"
	hostPlayerId: string;
	worldSeed?: number;
}
```

## 速率限制

- 每秒消息上限: `100` (可在 `config.yml` 调整)
- 超出限制时服务端静默丢弃消息
- 心跳 ping 不受速率限制
