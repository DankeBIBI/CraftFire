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
