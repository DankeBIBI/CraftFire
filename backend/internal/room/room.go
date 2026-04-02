package room

import (
	"fmt"
	"sync"
	"time"

	"CraftFire/backend/internal/config"
	applogger "CraftFire/backend/internal/logger"
	"CraftFire/backend/internal/player"
	"CraftFire/backend/internal/websocket"
)

// Room 代表一个游戏房间实例。
// 每个房间持有一个独立的 WebSocket Hub 和 Server。
type Room struct {
	ID         string
	Port       int
	MaxPlayers int
	GameMode   string
	CreatedAt  time.Time
	IsPublic   bool
	IsLocked   bool   // 房间是否锁定
	WorldSeed  string // 世界种子

	hub      *websocket.Hub
	wsServer *websocket.Server
	players  map[string]*player.State
	blocks   map[string]string // "x,y,z" -> blockType
	mu       sync.RWMutex

	// 统计数据
	TotalBlocksPlaced  int64
	TotalBlocksRemoved int64
	PeakPlayerCount    int
	TotalPlayersJoined int
}

// NewRoom 创建一个新的房间实例。
func NewRoom(id string, port int, cfg *config.Config, hub *websocket.Hub, wsServer *websocket.Server) *Room {
	return &Room{
		ID:         id,
		Port:       port,
		MaxPlayers: 50,
		GameMode:   "sandbox",
		CreatedAt:  time.Now(),
		IsPublic:   true,
		hub:        hub,
		wsServer:   wsServer,
		players:    make(map[string]*player.State),
		blocks:     make(map[string]string),
	}
}

// GetPlayerCount 返回当前在线玩家数量。
func (r *Room) GetPlayerCount() int {
	return r.hub.GetClientCount()
}

// GetPlayerInfoList 获取所有在线玩家的信息列表。
func (r *Room) GetPlayerInfoList() []player.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]player.Info, 0, len(r.players))
	for _, p := range r.players {
		list = append(list, player.Info{
			ID:             p.ID,
			Name:           p.Name,
			Position:       p.Position,
			Health:         p.Health,
			Status:         p.GetStatus(),
			ConnectedAt:    p.ConnectedAt.UnixMilli(),
			LastActivityAt: p.LastActivityAt.UnixMilli(),
			Ping:           p.Ping,
			Equipment:      p.Equipment,
		})
	}
	return list
}

// GetPlayerDetails 获取指定玩家的详细信息。
func (r *Room) GetPlayerDetails(playerId string) (*player.Details, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.players[playerId]
	if !exists {
		return nil, fmt.Errorf("玩家 %s 不在房间 %s 中", playerId, r.ID)
	}

	return p.ToDetails(), nil
}

// PlaceBlock 在房间世界中放置方块。
func (r *Room) PlaceBlock(x, y, z int, blockType string) error {
	key := fmt.Sprintf("%d,%d,%d", x, y, z)

	r.mu.Lock()
	r.blocks[key] = blockType
	r.TotalBlocksPlaced++
	r.mu.Unlock()

	// 广播世界更新
	msg, _ := websocket.NewMessage(websocket.MsgWorldUpdate, "", r.ID, websocket.WorldUpdatePayload{
		Changes: []websocket.WorldChangeEntry{{
			X:         x,
			Y:         y,
			Z:         z,
			BlockType: blockType,
			Action:    "place",
		}},
	})
	r.hub.Broadcast(msg)

	return nil
}

// RemoveBlock 从房间世界中移除方块。
func (r *Room) RemoveBlock(x, y, z int) error {
	key := fmt.Sprintf("%d,%d,%d", x, y, z)

	r.mu.Lock()
	delete(r.blocks, key)
	r.TotalBlocksRemoved++
	r.mu.Unlock()

	msg, _ := websocket.NewMessage(websocket.MsgWorldUpdate, "", r.ID, websocket.WorldUpdatePayload{
		Changes: []websocket.WorldChangeEntry{{
			X:      x,
			Y:      y,
			Z:      z,
			Action: "remove",
		}},
	})
	r.hub.Broadcast(msg)

	return nil
}

// KickPlayer 踢除指定玩家。
func (r *Room) KickPlayer(playerId string, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.players[playerId]
	if !exists {
		return fmt.Errorf("玩家 %s 不存在", playerId)
	}

	// 广播玩家离开消息
	msg, _ := websocket.NewMessage(websocket.MsgPlayerLeave, playerId, r.ID, websocket.PlayerLeavePayload{
		PlayerID: playerId,
	})
	r.hub.Broadcast(msg)

	delete(r.players, playerId)
	applogger.Info("玩家 %s 已被踢出房间 %s，原因: %s", playerId, r.ID, reason)
	return nil
}

// MutePlayer 禁言指定玩家。
func (r *Room) MutePlayer(playerId string, durationSeconds int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.players[playerId]
	if !exists {
		return fmt.Errorf("玩家 %s 不存在", playerId)
	}

	p.IsMuted = true
	p.MuteEndTime = time.Now().Add(time.Duration(durationSeconds) * time.Second)
	applogger.Info("玩家 %s 已被禁言 %d 秒", playerId, durationSeconds)
	return nil
}

// Shutdown 关闭房间，停止 WebSocket 服务器并清理资源。
func (r *Room) Shutdown() {
	applogger.Info("房间 %s 正在关闭", r.ID)
	if r.wsServer != nil {
		r.wsServer.Stop()
	}
}

// SetLocked 设置房间锁定状态。
func (r *Room) SetLocked(locked bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.IsLocked = locked
	applogger.Info("房间 %s 已%s", r.ID, map[bool]string{true: "锁定", false: "解锁"}[locked])
}

// Broadcast 广播消息给房间内所有玩家。
func (r *Room) Broadcast(message string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, err := websocket.NewMessage(websocket.MsgChat, "", r.ID, websocket.ChatPayload{
		PlayerID:   "SYSTEM",
		PlayerName: "系统",
		Content:    message,
		Timestamp:  time.Now().UnixMilli(),
	})
	if err != nil {
		return err
	}
	r.hub.Broadcast(msg)
	return nil
}

// HealPlayer 为指定玩家恢复生命值。
func (r *Room) HealPlayer(playerId string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.players[playerId]
	if !exists {
		return fmt.Errorf("玩家 %s 不存在", playerId)
	}

	p.Health = 100
	applogger.Info("玩家 %s 生命值已恢复", playerId)

	// 广播玩家状态更新
	msg, _ := websocket.NewMessage(websocket.MsgPlayerUpdate, playerId, r.ID, websocket.PlayerUpdatePayload{
		PlayerID: playerId,
		Health:   p.Health,
		Position: p.Position,
	})
	r.hub.Broadcast(msg)
	return nil
}

// TeleportPlayer 传送玩家到指定位置。
func (r *Room) TeleportPlayer(playerId string, x, y, z float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.players[playerId]
	if !exists {
		return fmt.Errorf("玩家 %s 不存在", playerId)
	}

	p.Position.X = x
	p.Position.Y = y
	p.Position.Z = z
	applogger.Info("玩家 %s 已被传送到 (%.1f, %.1f, %.1f)", playerId, x, y, z)

	// 广播玩家状态更新
	msg, _ := websocket.NewMessage(websocket.MsgPlayerUpdate, playerId, r.ID, websocket.PlayerUpdatePayload{
		PlayerID: playerId,
		Health:   p.Health,
		Position: p.Position,
	})
	r.hub.Broadcast(msg)
	return nil
}

// UpdateConfig 更新房间配置。
func (r *Room) UpdateConfig(maxPlayers int, gameMode, worldSeed string, isPublic bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if maxPlayers > 0 && maxPlayers <= 50 {
		r.MaxPlayers = maxPlayers
	}
	if gameMode != "" {
		r.GameMode = gameMode
	}
	if worldSeed != "" {
		r.WorldSeed = worldSeed
	}
	r.IsPublic = isPublic

	applogger.Info("房间 %s 配置已更新: maxPlayers=%d, gameMode=%s, isPublic=%v", r.ID, r.MaxPlayers, r.GameMode, r.IsPublic)
	return nil
}
