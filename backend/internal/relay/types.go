package relay

import (
	"encoding/json"
	"sync"
	"time"
)

// ─── 房间结构 ───────────────────────────────────────

// Room 互联网模式下的虚拟房间。
// 与 local/room 不同，relay 房间不启动独立端口，
// 而是托管在 relay 服务器上，通过中央 Hub 转发消息。
type Room struct {
	ID           string
	MaxPlayers   int
	HostID       string
	GameMode     string
	Players      map[string]*Client
	CreatedAt    time.Time
	LastActive   time.Time
	mu           sync.RWMutex
}

// NewRoom 创建一个新的 relay 房间。
func NewRoom(id, hostID string, maxPlayers int, gameMode string) *Room {
	return &Room{
		ID:         id,
		MaxPlayers: maxPlayers,
		HostID:     hostID,
		GameMode:   gameMode,
		Players:    make(map[string]*Client),
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
}

// AddPlayer 将玩家加入房间。
func (r *Room) AddPlayer(c *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= r.MaxPlayers {
		return false
	}

	r.Players[c.ID] = c
	r.LastActive = time.Now()
	return true
}

// RemovePlayer 将玩家从房间移除。
func (r *Room) RemovePlayer(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, clientID)
	r.LastActive = time.Now()
}

// PlayerCount 返回当前玩家数。
func (r *Room) PlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Players)
}

// MarshalJSON 房间信息序列化（不含玩家详情）。
func (r *Room) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type info struct {
		ID           string `json:"id"`
		MaxPlayers   int    `json:"maxPlayers"`
		PlayerCount  int    `json:"playerCount"`
		HostID       string `json:"hostId"`
		GameMode     string `json:"gameMode"`
		CreatedAt    int64  `json:"createdAt"`
	}

	return json.Marshal(info{
		ID:          r.ID,
		MaxPlayers:  r.MaxPlayers,
		PlayerCount: len(r.Players),
		HostID:      r.HostID,
		GameMode:    r.GameMode,
		CreatedAt:   r.CreatedAt.UnixMilli(),
	})
}

// ─── 客户端连接 ───────────────────────────────────────

// Client relay 服务器上的一个玩家连接。
type Client struct {
	ID        string
	Name      string
	RoomID    string
	ConnectedAt time.Time
	Hub       *Hub
	Send      chan []byte
}

// NewClient 创建一个新的 relay 客户端。
func NewClient(id, name, roomID string, hub *Hub) *Client {
	return &Client{
		ID:         id,
		Name:       name,
		RoomID:     roomID,
		ConnectedAt: time.Now(),
		Hub:        hub,
		Send:       make(chan []byte, 256),
	}
}
