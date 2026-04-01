package websocket

import (
	"time"

	applogger "CraftFire/backend/internal/logger"

	"github.com/gorilla/websocket"
)

const (
	// writeWait 写入操作的超时时间
	writeWait = 10 * time.Second

	// pongWait 等待 pong 响应的超时时间
	pongWait = 60 * time.Second

	// pingPeriod 发送 ping 的间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize 允许的最大消息大小（字节）
	maxMessageSize = 65536
)

// Client 代表一个 WebSocket 客户端连接。
// 每个加入房间的玩家对应一个 Client 实例。
type Client struct {
	// 客户端唯一标识（通常为玩家 UUID）
	ID string

	// 玩家昵称
	Name string

	// 所属房间号
	RoomID string

	// 所属 Hub
	hub *Hub

	// 底层 WebSocket 连接
	conn *websocket.Conn

	// 发送缓冲通道
	send chan []byte

	// 连接建立时间
	ConnectedAt time.Time

	// 最后活动时间
	LastActivity time.Time

	// 网络延迟（毫秒）
	Ping int64

	// 远程 IP 地址
	RemoteAddr string
}

// NewClient 创建一个新的客户端实例。
//
// 参数：
//   - hub: 所属的 Hub 实例
//   - conn: WebSocket 连接
//   - id: 客户端 ID
//   - name: 玩家昵称
//   - roomID: 房间号
func NewClient(hub *Hub, conn *websocket.Conn, id, name, roomID string) *Client {
	return &Client{
		ID:           id,
		Name:         name,
		RoomID:       roomID,
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, 256),
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		RemoteAddr:   conn.RemoteAddr().String(),
	}
}

// ReadPump 持续从 WebSocket 连接读取消息并转发给 Hub。
// 在独立的 goroutine 中运行，客户端断开时自动注销。
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				applogger.Warn("客户端 %s WebSocket 异常关闭: %v", c.ID, err)
			}
			break
		}
		c.LastActivity = time.Now()
		// 通过消息处理器处理消息，而不是直接广播
		HandleMessage(c.hub, c, message)
	}
}

// WritePump 持续将消息从发送通道写入 WebSocket 连接。
// 同时负责发送心跳 ping 帧。
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送排队中的消息（减少系统调用）
			n := len(c.send)
			for range n {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send 向客户端发送消息。
func (c *Client) Send(message []byte) {
	select {
	case c.send <- message:
	default:
		applogger.Warn("客户端 %s 发送缓冲区满，丢弃消息", c.ID)
	}
}
