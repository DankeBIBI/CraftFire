package relay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	applogger "CraftFire/backend/internal/logger"

	"github.com/gorilla/websocket"
)

// ─── HTTP Server ─────────────────────────────────────

// Server HTTP + WebSocket 服务器。
// 提供 REST API（房间管理）和 WebSocket 中继。
type Server struct {
	hub      *Hub
	addr     string
	mux      *http.ServeMux
	server   *http.Server
	upgrader websocket.Upgrader
}

// NewServer 创建一个新的 relay 服务器。
func NewServer(addr string) *Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true // 生产环境应限制来源
		},
	}

	s := &Server{
		hub:      NewHub(),
		addr:     addr,
		mux:      http.NewServeMux(),
		upgrader: upgrader,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 注册路由。
func (s *Server) setupRoutes() {
	// REST API
	s.mux.HandleFunc("GET  /api/rooms", s.handleListRooms)
	s.mux.HandleFunc("POST /api/rooms", s.handleCreateRoom)
	s.mux.HandleFunc("GET  /api/rooms/", s.handleGetRoom)
	s.mux.HandleFunc("DELETE /api/rooms/", s.handleDeleteRoom)

	// WebSocket 中继
	s.mux.HandleFunc("GET  /ws", s.handleWebSocket)

	// 健康检查
	s.mux.HandleFunc("GET  /health", s.handleHealth)
}

// Start 启动 relay 服务器。
func (s *Server) Start() error {
	// 启动 Hub
	go s.hub.Run()

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	applogger.Info("[Relay] 服务器启动于 %s", s.addr)
	return s.server.ListenAndServe()
}

// Stop 停止 relay 服务器。
func (s *Server) Stop() error {
	s.hub.Stop()
	return s.server.Close()
}

// ─── REST API Handlers ──────────────────────────────

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	rooms := s.hub.ListRooms()
	data := make([]json.RawMessage, 0, len(rooms))
	for _, room := range rooms {
		b, _ := room.MarshalJSON()
		data = append(data, b)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rooms": data,
		"total": len(data),
	})
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		HostID     string `json:"hostId"`
		MaxPlayers int    `json:"maxPlayers"`
		GameMode   string `json:"gameMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.MaxPlayers <= 0 || req.MaxPlayers > 50 {
		req.MaxPlayers = 10
	}
	if req.GameMode == "" {
		req.GameMode = "sandbox"
	}

	roomID, err := s.hub.CreateRoom(req.HostID, req.MaxPlayers, req.GameMode)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	room := s.hub.GetRoom(roomID)
	b, _ := room.MarshalJSON()
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"room": json.RawMessage(b),
	})
}

func (s *Server) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := extractRoomID(r.URL.Path)
	if roomID == "" || !roomIDPattern.MatchString(roomID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid room id"})
		return
	}

	room := s.hub.GetRoom(roomID)
	if room == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	b, _ := room.MarshalJSON()
	writeJSON(w, http.StatusOK, map[string]interface{}{"room": json.RawMessage(b)})
}

func (s *Server) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	// 仅允许房主删除
	var req struct {
		RoomID string `json:"roomId"`
		HostID string `json:"hostId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	room := s.hub.GetRoom(req.RoomID)
	if room == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
		return
	}

	if room.HostID != req.HostID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "not room host"})
		return
	}

	// 解散房间：断开所有连接
	room.mu.RLock()
	for _, c := range room.Players {
		s.hub.UnregisterClient(c)
	}
	room.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]string{"message": "room deleted"})
}

// extractRoomID 从路径中提取房间号。
func extractRoomID(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/api/rooms/"), "/")
	if len(parts) > 0 && roomIDPattern.MatchString(parts[0]) {
		return parts[0]
	}
	return ""
}

// ─── WebSocket Handler ──────────────────────────────

// wsClient WebSocket 连接包装。
type wsClient struct {
	hub    *Hub
	conn   *websocket.Conn
	client *Client
	done   chan struct{}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		applogger.Error("[Relay] WebSocket 升级失败: %v", err)
		return
	}

	// 解析参数
	roomID := r.URL.Query().Get("room")
	playerName := r.URL.Query().Get("name")
	action := r.URL.Query().Get("action") // create | join | leave

	if playerName == "" {
		playerName = "玩家"
	}

	clientID := generateID()
	c := NewClient(clientID, playerName, roomID, s.hub)
	wsc := &wsClient{hub: s.hub, conn: conn, client: c, done: make(chan struct{})}

	// 房主创建房间
	if action == "create" {
		maxPlayers, _ := strconv.Atoi(r.URL.Query().Get("max"))
		if maxPlayers <= 0 {
			maxPlayers = 10
		}
		gameMode := r.URL.Query().Get("mode")
		if gameMode == "" {
			gameMode = "sandbox"
		}
		rid, err := s.hub.CreateRoom(clientID, maxPlayers, gameMode)
		if err != nil {
			conn.WriteJSON(map[string]string{"type": "error", "error": err.Error()})
			conn.Close()
			return
		}
		c.RoomID = rid
		conn.WriteJSON(map[string]interface{}{
			"type":   "room_created",
			"roomId": rid,
			"hostId": clientID,
		})
		applogger.Info("[Relay] 房主 %s 创建房间 %s", playerName, rid)
	}

	// 加入房间
	if roomID != "" && roomIDPattern.MatchString(roomID) {
		if !s.hub.JoinRoom(roomID, c) {
			conn.WriteJSON(map[string]string{"type": "error", "error": "room full or not found"})
			conn.Close()
			return
		}
		// 广播加入
		joinMsg, _ := json.Marshal(map[string]interface{}{
			"type":       "player_join",
			"playerId":   clientID,
			"playerName": playerName,
			"roomId":     roomID,
		})
		s.hub.BroadcastRoom(roomID, joinMsg, clientID)
	}

	s.hub.RegisterClient(c)

	// 读写
	go wsc.writePump()
	wsc.readPump()
}

func (c *wsClient) readPump() {
	defer func() {
		c.hub.UnregisterClient(c.client)
		close(c.done)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// 透传消息到房间
		if c.client.RoomID != "" {
			// 解析类型标记（不改动消息体）
			var msg struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &msg) == nil && msg.Type != "" {
				// 注入 playerId
				c.hub.BroadcastRoom(c.client.RoomID, data, c.client.ID)
			}
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.client.Send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	rooms := s.hub.ListRooms()
	totalPlayers := 0
	for _, r := range rooms {
		totalPlayers += r.PlayerCount()
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"rooms":   len(rooms),
		"players": totalPlayers,
	})
}

// ─── 辅助 ──────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func generateID() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}
