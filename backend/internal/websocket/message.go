// Package websocket 提供 CraftFire 的 WebSocket 通信核心。
// 包含 Hub（消息中枢）、Client（客户端连接）和 Server（WS 服务器）。
package websocket

import (
	"encoding/json"
	"time"
)

// MessageType 消息类型常量定义
const (
	MsgPlayerJoin      = "player_join"
	MsgPlayerLeave     = "player_leave"
	MsgPlayerMove      = "player_move"
	MsgPlayerEquip     = "player_equip"
	MsgPlayerStateSync = "player_state_sync"
	MsgPlayerUpdate    = "player_update"
	MsgBlockPlace      = "block_place"
	MsgBlockRemove     = "block_remove"
	MsgWorldUpdate     = "world_update"
	MsgWorldSnapshot   = "world_snapshot"
	MsgPing            = "ping"
	MsgPong            = "pong"
	MsgChat            = "chat"
	MsgError           = "error"
	MsgBroadcast       = "broadcast"
)

// Message WebSocket 消息信封结构体。
// 所有 WebSocket 消息均使用此统一格式传输。
type Message struct {
	Type      string          `json:"type"`      // 消息类型
	Timestamp int64           `json:"timestamp"` // 毫秒级时间戳
	PlayerID  string          `json:"playerId"`  // 发送者玩家 ID
	RoomID    string          `json:"roomId"`    // 房间号
	ID        string          `json:"id"`        // 消息唯一标识
	Payload   json.RawMessage `json:"payload"`   // 具体数据
}

// NewMessage 创建一个新的消息实例。
//
// 参数：
//   - msgType: 消息类型（使用 Msg* 常量）
//   - playerID: 发送者 ID
//   - roomID: 房间号
//   - payload: 消息负载数据
//
// 返回值：序列化后的 JSON 字节切片
func NewMessage(msgType, playerID, roomID string, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Type:      msgType,
		Timestamp: time.Now().UnixMilli(),
		PlayerID:  playerID,
		RoomID:    roomID,
		Payload:   payloadBytes,
	}

	return json.Marshal(msg)
}

// ParseMessage 解析原始 JSON 字节为 Message 结构体。
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// PlayerMovePayload 玩家移动消息的负载数据。
type PlayerMovePayload struct {
	X        float64         `json:"x"`
	Y        float64         `json:"y"`
	Z        float64         `json:"z"`
	Rotation RotationPayload `json:"rotation"`
}

// RotationPayload 旋转数据。
type RotationPayload struct {
	Pitch float64 `json:"pitch"`
	Yaw   float64 `json:"yaw"`
}

// BlockPlacePayload 放置方块消息的负载数据。
type BlockPlacePayload struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Z         int    `json:"z"`
	BlockType string `json:"blockType"`
}

// BlockRemovePayload 移除方块消息的负载数据。
type BlockRemovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// PlayerJoinPayload 玩家加入消息的负载数据。
type PlayerJoinPayload struct {
	PlayerID   string  `json:"playerId"`
	PlayerName string  `json:"playerName"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
	Equipment  string  `json:"equipment"`
}

// PlayerLeavePayload 玩家离开消息的负载数据。
type PlayerLeavePayload struct {
	PlayerID string `json:"playerId"`
}

// PlayerEquipPayload 玩家装备切换消息的负载数据。
type PlayerEquipPayload struct {
	PlayerID  string `json:"playerId"`
	Equipment string `json:"equipment"`
}

// WorldChangeEntry 世界变化条目。
type WorldChangeEntry struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Z         int    `json:"z"`
	BlockType string `json:"blockType"`
	Action    string `json:"action"` // "place" 或 "remove"
}

// WorldUpdatePayload 世界更新消息的负载数据。
type WorldUpdatePayload struct {
	Changes []WorldChangeEntry `json:"changes"`
}

// ChatPayload 聊天消息的负载数据。
type ChatPayload struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

// PlayerUpdatePayload 玩家状态更新消息的负载数据。
type PlayerUpdatePayload struct {
	PlayerID string `json:"playerId"`
	Health   int    `json:"health"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"position"`
}

// BroadcastPayload 广播消息的负载数据。
type BroadcastPayload struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
