package websocket

import (
	"fmt"
	"net/http"

	applogger "CraftFire/backend/internal/logger"
	"CraftFire/backend/internal/utils"

	"github.com/gorilla/websocket"
)

// upgrader 用于将 HTTP 连接升级为 WebSocket 连接。
// CheckOrigin 在局域网场景下允许所有来源。
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // 局域网环境允许所有来源
	},
}

// Server 是 WebSocket 服务器，用于接受客户端连接。
// 每个房间持有一个 Server 实例，监听不同端口。
type Server struct {
	hub    *Hub
	port   int
	roomID string
	server *http.Server
}

// NewServer 创建一个新的 WebSocket 服务器实例。
//
// 参数：
//   - hub: 消息中枢
//   - port: 监听端口（使用房间号作为端口号）
//   - roomID: 房间号
func NewServer(hub *Hub, port int, roomID string) *Server {
	return &Server{
		hub:    hub,
		port:   port,
		roomID: roomID,
	}
}

// Start 启动 WebSocket 服务器，开始监听指定端口。
// 应在独立的 goroutine 中运行。
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	applogger.Info("WebSocket 服务器启动于端口 %d (房间 %s)", s.port, s.roomID)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("WebSocket 服务器启动失败: %w", err)
	}
	return nil
}

// Stop 优雅停止 WebSocket 服务器。
func (s *Server) Stop() error {
	if s.server != nil {
		applogger.Info("正在停止 WebSocket 服务器 (端口 %d)", s.port)
		s.hub.Stop()
		return s.server.Close()
	}
	return nil
}

// handleWebSocket 处理 WebSocket 连接升级请求。
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		applogger.Error("WebSocket 升级失败: %v", err)
		return
	}

	// 从查询参数获取玩家信息
	playerName := r.URL.Query().Get("name")
	if playerName == "" {
		playerName = "未知玩家"
	}

	playerID := utils.GenerateUUID()

	client := NewClient(s.hub, conn, playerID, playerName, s.roomID)
	s.hub.Register(client)

	applogger.Info("玩家 %s (%s) 已连接到房间 %s", playerName, playerID, s.roomID)

	// 启动读写泵
	go client.WritePump()
	go client.ReadPump()

	// 广播玩家加入消息
	joinMsg, _ := NewMessage(MsgPlayerJoin, playerID, s.roomID, PlayerJoinPayload{
		PlayerID:   playerID,
		PlayerName: playerName,
		X:          0,
		Y:          10,
		Z:          0,
		Equipment:  "pistol", // 默认装备
	})
	s.hub.Broadcast(joinMsg)
}

// handleHealth 健康检查端点。
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","room":"%s","players":%d}`, s.roomID, s.hub.GetClientCount())
}
