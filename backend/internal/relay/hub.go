package relay

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sync"
	"time"

	applogger "CraftFire/backend/internal/logger"
)

// ─── Relay Hub ───────────────────────────────────────

// Hub 是 relay 服务器的中央消息中枢。
// 管理所有房间和客户端连接，负责消息路由。
type Hub struct {
	rooms   map[string]*Room   // roomID -> Room
	clients map[string]*Client // clientID -> Client
	mu      sync.RWMutex

	// 通道
	register   chan *Client
	unregister chan *Client
	broadcast  chan *relayMessage

	done chan struct{}
}

// relayMessage 内部广播消息结构。
type relayMessage struct {
	RoomID    string
	ExcludeID string
	Data      []byte
}

// NewHub 创建一个新的 relay Hub。
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *relayMessage, 512),
		done:       make(chan struct{}),
	}
}

// Run 启动 Hub 的主事件循环。
func (h *Hub) Run() {
	applogger.Info("[Relay] Hub 已启动")
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.ID] = c
			h.mu.Unlock()
			applogger.Info("[Relay] 客户端注册: %s (%s)", c.Name, c.ID)

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c.ID]; ok {
				// 从房间移除
				if r := h.rooms[c.RoomID]; r != nil {
					r.RemovePlayer(c.ID)
					// 通知房间内其他玩家有人离开
					h.broadcast <- &relayMessage{
						RoomID:    c.RoomID,
						ExcludeID: c.ID,
						Data:      newSystemMsg(c.RoomID, "player_leave", map[string]string{"playerId": c.ID}),
					}
					// 房主离开则解散房间
					if r.HostID == c.ID && r.PlayerCount() == 0 {
						delete(h.rooms, c.RoomID)
						applogger.Info("[Relay] 房间 %s 已解散（房主离开且无玩家）", c.RoomID)
					}
				}
				delete(h.clients, c.ID)
				close(c.Send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[msg.RoomID]
			if room == nil {
				h.mu.RUnlock()
				continue
			}
			room.mu.RLock()
			for _, c := range room.Players {
				if msg.ExcludeID != "" && c.ID == msg.ExcludeID {
					continue
				}
				select {
				case c.Send <- msg.Data:
				default:
					// 缓冲区满，关闭
					h.unregister <- c
				}
			}
			room.mu.RUnlock()
			h.mu.RUnlock()

		case <-h.done:
			applogger.Info("[Relay] Hub 正在关闭")
			return
		}
	}
}

// RegisterClient 注册客户端到 Hub。
func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}

// UnregisterClient 从 Hub 注销客户端。
func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}

// BroadcastRoom 向房间内所有玩家广播消息（可排除指定玩家）。
func (h *Hub) BroadcastRoom(roomID string, data []byte, excludeID string) {
	h.broadcast <- &relayMessage{RoomID: roomID, ExcludeID: excludeID, Data: data}
}

// SendToClient 向指定玩家发送消息。
func (h *Hub) SendToClient(clientID string, data []byte) bool {
	h.mu.RLock()
	c, ok := h.clients[clientID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case c.Send <- data:
		return true
	default:
		return false
	}
}

// ─── 房间管理 ───────────────────────────────────────

var roomIDPattern = regexp.MustCompile(`^\d{6}$`)

// CreateRoom 创建一个新的 relay 房间。
// 生成 6 位房间号，与本地模式保持一致的接口。
func (h *Hub) CreateRoom(hostID string, maxPlayers int, gameMode string) (string, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	roomID := fmt.Sprintf("%06d", 100000+rng.Intn(900000))

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[roomID]; exists {
		return "", fmt.Errorf("房间号冲突")
	}

	room := NewRoom(roomID, hostID, maxPlayers, gameMode)
	h.rooms[roomID] = room
	applogger.Info("[Relay] 房间已创建: %s (房主: %s)", roomID, hostID)
	return roomID, nil
}

// JoinRoom 将玩家加入 relay 房间。
func (h *Hub) JoinRoom(roomID string, c *Client) bool {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()
	if !ok {
		return false
	}

	added := room.AddPlayer(c)
	if added {
		c.RoomID = roomID
		applogger.Info("[Relay] %s (%s) 加入房间 %s", c.Name, c.ID, roomID)
	}
	return added
}

// LeaveRoom 让玩家离开当前房间。
func (h *Hub) LeaveRoom(c *Client) {
	h.mu.RLock()
	room := h.rooms[c.RoomID]
	h.mu.RUnlock()

	if room != nil {
		room.RemovePlayer(c.ID)
		// 广播离开消息
		h.BroadcastRoom(room.ID, newSystemMsg(room.ID, "player_leave", map[string]string{"playerId": c.ID}), c.ID)
	}
}

// ListRooms 返回所有活跃房间列表。
func (h *Hub) ListRooms() []*Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := make([]*Room, 0, len(h.rooms))
	for _, r := range h.rooms {
		list = append(list, r)
	}
	return list
}

// GetRoom 返回指定房间信息。
func (h *Hub) GetRoom(roomID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

// ─── 辅助函数 ───────────────────────────────────────

// newSystemMsg 构造系统消息。
func newSystemMsg(roomID, msgType string, payload interface{}) []byte {
	payloadBytes, _ := json.Marshal(payload)
	return json.RawMessage(fmt.Sprintf(
		`{"type":"%s","timestamp":%d,"roomId":"%s","payload":%s}`,
		msgType, time.Now().UnixMilli(), roomID, string(payloadBytes),
	))
}

// Stop 停止 Hub。
func (h *Hub) Stop() {
	close(h.done)
}
