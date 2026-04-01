package websocket

import (
	"encoding/json"
	"math"

	applogger "CraftFire/backend/internal/logger"
)

// 坐标合法性边界
const (
	MinCoord = -1000.0
	MaxCoord = 1000.0
)

// HandleMessage 路由和处理收到的 WebSocket 消息。
// 根据消息类型分发到对应的处理函数。
//
// 参数：
//   - hub: Hub 实例
//   - client: 发送消息的客户端
//   - rawMessage: 原始消息字节
func HandleMessage(hub *Hub, client *Client, rawMessage []byte) {
	msg, err := ParseMessage(rawMessage)
	if err != nil {
		applogger.Warn("无法解析来自 %s 的消息: %v", client.ID, err)
		return
	}

	switch msg.Type {
	case MsgPlayerMove:
		handlePlayerMove(hub, client, msg)
	case MsgPlayerEquip:
		handlePlayerEquip(hub, client, msg)
	case MsgBlockPlace:
		handleBlockPlace(hub, client, msg)
	case MsgBlockRemove:
		handleBlockRemove(hub, client, msg)
	case MsgPing:
		handlePing(client, msg)
	case MsgChat:
		handleChat(hub, client, msg)
	default:
		applogger.Debug("未知消息类型: %s (来自 %s)", msg.Type, client.ID)
	}
}

// handlePlayerMove 处理玩家移动消息，转发给其他客户端。
func handlePlayerMove(hub *Hub, client *Client, msg *Message) {
	var payload PlayerMovePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		applogger.Warn("无法解析玩家移动数据: %v", err)
		return
	}

	// 校验坐标合法性
	if !isValidCoord(payload.X) || !isValidCoord(payload.Y) || !isValidCoord(payload.Z) {
		applogger.Warn("非法坐标来自 %s: %.2f, %.2f, %.2f", client.ID, payload.X, payload.Y, payload.Z)
		return
	}

	// 校验旋转角度
	if !isValidAngle(payload.Rotation.Pitch) || !isValidAngle(payload.Rotation.Yaw) {
		return
	}

	// 将移动数据广播给其他玩家
	broadcastMsg, _ := NewMessage(MsgPlayerMove, client.ID, client.RoomID, payload)
	hub.BroadcastExcept(broadcastMsg, client.ID)
}

// handlePlayerEquip 处理玩家装备切换消息，广播给其他客户端。
func handlePlayerEquip(hub *Hub, client *Client, msg *Message) {
	var payload PlayerEquipPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		applogger.Warn("无法解析玩家装备数据: %v", err)
		return
	}

	// 校验装备标识
	if payload.Equipment == "" || len(payload.Equipment) > 32 {
		applogger.Warn("非法装备标识来自 %s: %s", client.ID, payload.Equipment)
		return
	}

	// 广播装备更新给其他玩家
	broadcastMsg, _ := NewMessage(MsgPlayerEquip, client.ID, client.RoomID, payload)
	hub.BroadcastExcept(broadcastMsg, client.ID)
}

// handleBlockPlace 处理方块放置消息。
func handleBlockPlace(hub *Hub, client *Client, msg *Message) {
	var payload BlockPlacePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		applogger.Warn("无法解析方块放置数据: %v", err)
		return
	}

	// 校验坐标和方块类型
	if !isValidBlockCoord(payload.X, payload.Y, payload.Z) {
		applogger.Warn("非法方块坐标来自 %s: %d, %d, %d", client.ID, payload.X, payload.Y, payload.Z)
		return
	}

	if payload.BlockType == "" || len(payload.BlockType) > 32 {
		applogger.Warn("非法方块类型来自 %s: %s", client.ID, payload.BlockType)
		return
	}

	// 广播世界更新
	worldUpdate, _ := NewMessage(MsgWorldUpdate, client.ID, client.RoomID, WorldUpdatePayload{
		Changes: []WorldChangeEntry{{
			X:         payload.X,
			Y:         payload.Y,
			Z:         payload.Z,
			BlockType: payload.BlockType,
			Action:    "place",
		}},
	})
	hub.Broadcast(worldUpdate)
}

// handleBlockRemove 处理方块移除消息。
func handleBlockRemove(hub *Hub, client *Client, msg *Message) {
	var payload BlockRemovePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		applogger.Warn("无法解析方块移除数据: %v", err)
		return
	}

	// 校验坐标
	if !isValidBlockCoord(payload.X, payload.Y, payload.Z) {
		applogger.Warn("非法方块坐标来自 %s: %d, %d, %d", client.ID, payload.X, payload.Y, payload.Z)
		return
	}

	worldUpdate, _ := NewMessage(MsgWorldUpdate, client.ID, client.RoomID, WorldUpdatePayload{
		Changes: []WorldChangeEntry{{
			X:      payload.X,
			Y:      payload.Y,
			Z:      payload.Z,
			Action: "remove",
		}},
	})
	hub.Broadcast(worldUpdate)
}

// handlePing 处理心跳检测，回复 pong。
func handlePing(client *Client, _ *Message) {
	pongMsg, _ := NewMessage(MsgPong, client.ID, client.RoomID, nil)
	client.Send(pongMsg)
}

// handleChat 处理聊天消息，广播给所有玩家。
func handleChat(hub *Hub, _ *Client, msg *Message) {
	// 直接转发聊天消息
	rawMsg, _ := json.Marshal(msg)
	hub.Broadcast(rawMsg)
}

// isValidCoord 检查坐标是否在合法范围内
func isValidCoord(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= MinCoord && v <= MaxCoord
}

// isValidAngle 检查角度是否在合法范围内
func isValidAngle(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= -math.Pi*2 && v <= math.Pi*2
}

// isValidBlockCoord 检查方块坐标是否在合法范围内
func isValidBlockCoord(x, y, z int) bool {
	return x >= -100 && x <= 100 && y >= 0 && y <= 50 && z >= -100 && z <= 100
}
