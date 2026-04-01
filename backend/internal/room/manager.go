// Package room 提供游戏房间的创建、管理和销毁功能。
// 每个房间对应一个独立的 WebSocket 服务器。
package room

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"sync"
	"time"

	"CraftFire/backend/internal/config"
	applogger "CraftFire/backend/internal/logger"
	"CraftFire/backend/internal/player"
	"CraftFire/backend/internal/websocket"
)

// Manager 房间管理器，负责游戏房间的全生命周期管理。
type Manager struct {
	rooms       map[string]*Room
	mu          sync.RWMutex
	cfg         *config.Config
	hub         *websocket.Hub
	currentRoom string // 当前加入的房间号
}

// NewManager 创建一个新的房间管理器。
func NewManager(cfg *config.Config, hub *websocket.Hub) *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
		cfg:   cfg,
		hub:   hub,
	}
}

// CreateRoom 创建一个新的游戏房间。
// 生成 6 位随机房间号，并在对应端口上启动 WebSocket 服务器。
//
// 返回值：
//   - roomId: 6 位数字房间号
//   - error: 创建失败时返回错误
func (m *Manager) CreateRoom() (string, error) {
	// 生成 6 位随机房间号（100000-999999）
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	roomID := strconv.Itoa(100000 + rng.Intn(900000))

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查房间号是否冲突
	if _, exists := m.rooms[roomID]; exists {
		return "", fmt.Errorf("房间号 %s 已存在", roomID)
	}

	// 检查是否超过最大房间数
	if len(m.rooms) >= m.cfg.Server.MaxRoomsPerIP {
		return "", fmt.Errorf("已达到最大房间数 (%d)", m.cfg.Server.MaxRoomsPerIP)
	}

	// 将 6 位房间号映射到合法端口范围 (10000-65535)
	roomNum, _ := strconv.Atoi(roomID)
	port := 10000 + (roomNum % 55536)

	// 创建房间专属的 Hub 和 WS 服务器
	roomHub := websocket.NewHub()
	wsServer := websocket.NewServer(roomHub, port, roomID)

	room := NewRoom(roomID, port, m.cfg, roomHub, wsServer)
	m.rooms[roomID] = room

	// 启动 Hub 和 WS 服务器
	go roomHub.Run()
	go func() {
		if err := wsServer.Start(); err != nil {
			applogger.Error("房间 %s 的 WebSocket 服务器启动失败: %v", roomID, err)
		}
	}()

	m.currentRoom = roomID
	applogger.Info("房间 %s 已创建 (端口 %d)", roomID, port)
	return roomID, nil
}

// JoinRoom 加入指定的游戏房间。
//
// 参数：
//   - roomId: 6 位数字房间号
//   - playerName: 玩家昵称
func (m *Manager) JoinRoom(roomId string, ip string) error {
	if !regexp.MustCompile(`^\d{6}$`).MatchString(roomId) {
		return fmt.Errorf("房间号必须为 6 位数字")
	}

	if ip != "" && ip != "127.0.0.1" && ip != "localhost" {
		m.currentRoom = roomId
		applogger.Info("加入远端房间 %s (%s)", roomId, ip)
		return nil
	}

	m.mu.RLock()
	r, exists := m.rooms[roomId]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("房间 %s 不存在", roomId)
	}

	if r.GetPlayerCount() >= r.MaxPlayers {
		return fmt.Errorf("房间 %s 已满", roomId)
	}

	m.currentRoom = roomId
	applogger.Info("加入本地房间 %s", roomId)
	return nil
}

// LeaveCurrentRoom 离开当前房间。
func (m *Manager) LeaveCurrentRoom() error {
	if m.currentRoom == "" {
		return fmt.Errorf("当前未在任何房间中")
	}

	applogger.Info("离开房间 %s", m.currentRoom)
	m.currentRoom = ""
	return nil
}

// GetRoom 获取指定房间的实例。
func (m *Manager) GetRoom(roomId string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomId]
	if !exists {
		return nil, fmt.Errorf("房间 %s 不存在", roomId)
	}
	return r, nil
}

// PlaceBlock 在当前房间放置方块。
func (m *Manager) PlaceBlock(x, y, z int, blockType string) error {
	if m.currentRoom == "" {
		return fmt.Errorf("未加入任何房间")
	}

	r, err := m.GetRoom(m.currentRoom)
	if err != nil {
		return err
	}

	return r.PlaceBlock(x, y, z, blockType)
}

// RemoveBlock 在当前房间移除方块。
func (m *Manager) RemoveBlock(x, y, z int) error {
	if m.currentRoom == "" {
		return fmt.Errorf("未加入任何房间")
	}

	r, err := m.GetRoom(m.currentRoom)
	if err != nil {
		return err
	}

	return r.RemoveBlock(x, y, z)
}

// ShutdownAll 关闭所有房间（应用退出时调用）。
func (m *Manager) ShutdownAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, r := range m.rooms {
		applogger.Info("正在关闭房间 %s", id)
		r.Shutdown()
		delete(m.rooms, id)
	}
}

// GetCurrentRoomID 返回当前已加入的房间号。
func (m *Manager) GetCurrentRoomID() string {
	return m.currentRoom
}

// GetRoomList 返回所有活跃房间的信息列表。
func (m *Manager) GetRoomList() []RoomInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]RoomInfo, 0, len(m.rooms))
	for _, r := range m.rooms {
		list = append(list, RoomInfo{
			RoomID:      r.ID,
			Port:        r.Port,
			PlayerCount: r.GetPlayerCount(),
			MaxPlayers:  r.MaxPlayers,
			CreatedAt:   r.CreatedAt.UnixMilli(),
		})
	}
	return list
}

// RoomInfo 房间基础信息（用于局域网发现）。
type RoomInfo struct {
	RoomID      string `json:"roomId"`
	Port        int    `json:"port"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	CreatedAt   int64  `json:"createdAt"`
}

// GetPlayerInfoFromRoom 获取指定房间内玩家信息列表。
func (m *Manager) GetPlayerInfoFromRoom(roomId string) ([]player.Info, error) {
	r, err := m.GetRoom(roomId)
	if err != nil {
		return nil, err
	}
	return r.GetPlayerInfoList(), nil
}

// 获取全部房间
func (m *Manager) GetAllRooms() []Room {
	rooms := make([]Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, *r)
	}
	return rooms
}
