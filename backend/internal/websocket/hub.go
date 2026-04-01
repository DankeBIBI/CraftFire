package websocket

import (
	"sync"
	"time"

	applogger "CraftFire/backend/internal/logger"
)

// Hub 是 WebSocket 消息中枢，负责管理所有客户端连接和消息广播。
// 每个游戏房间拥有一个独立的 Hub 实例。
type Hub struct {
	// 已注册的客户端集合
	clients map[*Client]bool

	// 广播通道：向所有客户端发送消息
	broadcast chan []byte

	// 注册通道：新客户端加入
	register chan *Client

	// 注销通道：客户端断开
	unregister chan *Client

	// 读写互斥锁，保护 clients map
	mu sync.RWMutex

	// 游戏循环 Tick Rate（毫秒）
	tickInterval time.Duration

	// 关闭信号
	done chan struct{}
}

// NewHub 创建一个新的 Hub 实例。
// 默认 Tick Rate 为 8 ticks/秒（125ms）。
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		tickInterval: 125 * time.Millisecond, // 8 ticks/sec
		done:         make(chan struct{}),
	}
}

// Run 启动 Hub 的主事件循环。
// 处理客户端注册、注销和消息广播。
// 应在独立的 goroutine 中运行。
func (h *Hub) Run() {
	applogger.Info("WebSocket Hub 已启动")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			applogger.Info("客户端已注册: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			applogger.Info("客户端已注销: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 发送缓冲区满，关闭客户端连接
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()

		case <-h.done:
			applogger.Info("WebSocket Hub 正在关闭")
			return
		}
	}
}

// Register 将客户端注册到 Hub。
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 从 Hub 注销客户端。
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast 向所有已连接的客户端广播消息。
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// BroadcastExcept 向除指定客户端外的所有客户端广播消息。
// 常用于将玩家操作转发给其他玩家。
func (h *Hub) BroadcastExcept(message []byte, excludeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.ID == excludeID {
			continue
		}
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// GetClientCount 返回当前已连接的客户端数量。
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetClientIDs 返回所有已连接客户端的 ID 列表。
func (h *Hub) GetClientIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.clients))
	for client := range h.clients {
		ids = append(ids, client.ID)
	}
	return ids
}

// Stop 停止 Hub 的事件循环并关闭所有客户端。
func (h *Hub) Stop() {
	close(h.done)

	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}
