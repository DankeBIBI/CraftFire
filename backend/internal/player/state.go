// Package player 提供玩家状态管理、数据结构和同步功能。
package player

import (
	"time"

	"CraftFire/backend/internal/utils"
)

// State 玩家完整状态，由后端权威维护。
type State struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Position       utils.Vec3     `json:"position"`
	Velocity       utils.Vec3     `json:"velocity"`
	Rotation       utils.Rotation `json:"rotation"`
	Health         int            `json:"health"`
	MaxHealth      int            `json:"maxHealth"`
	Ammo           int            `json:"ammo"`
	Equipment      string         `json:"equipment"`
	IsAlive        bool           `json:"isAlive"`
	LastUpdateTime int64          `json:"lastUpdateTime"`

	// 连接信息
	ConnectedAt    time.Time `json:"-"`
	LastActivityAt time.Time `json:"-"`
	RemoteIP       string    `json:"-"`
	Ping           int64     `json:"ping"`

	// 禁言信息
	IsMuted     bool      `json:"isMuted"`
	MuteEndTime time.Time `json:"-"`

	// 统计数据
	Statistics PlayerStats `json:"statistics"`

	// 背包
	Inventory []InventoryItem `json:"inventory"`
}

// PlayerStats 玩家游戏统计数据。
type PlayerStats struct {
	BlocksPlaced     int64   `json:"blocksPlaced"`
	BlocksRemoved    int64   `json:"blocksRemoved"`
	KillCount        int     `json:"killCount"`
	DeathCount       int     `json:"deathCount"`
	DistanceTraveled float64 `json:"distanceTraveled"`
}

// InventoryItem 背包物品。
type InventoryItem struct {
	ItemID   string      `json:"itemId"`
	ItemType string      `json:"itemType"`
	Quantity int         `json:"quantity"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// NewState 创建一个新的玩家状态实例。
func NewState(id, name string) *State {
	return &State{
		ID:             id,
		Name:           name,
		Position:       utils.Vec3{X: 0, Y: 10, Z: 0},
		Velocity:       utils.Vec3{},
		Rotation:       utils.Rotation{},
		Health:         100,
		MaxHealth:      100,
		Ammo:           30,
		Equipment:      "pistol",
		IsAlive:        true,
		LastUpdateTime: time.Now().UnixMilli(),
		ConnectedAt:    time.Now(),
		LastActivityAt: time.Now(),
		Inventory:      make([]InventoryItem, 0),
	}
}

// GetStatus 获取玩家状态字符串。
func (s *State) GetStatus() string {
	if !s.IsAlive {
		return "dead"
	}
	// 超过 5 分钟无活动视为闲置
	if time.Since(s.LastActivityAt) > 5*time.Minute {
		return "idle"
	}
	return "online"
}

// ToDetails 将 State 转换为完整的 Details 结构体。
func (s *State) ToDetails() *Details {
	return &Details{
		ID:        s.ID,
		Name:      s.Name,
		Position:  s.Position,
		Velocity:  s.Velocity,
		Rotation:  s.Rotation,
		Health:    s.Health,
		MaxHealth: s.MaxHealth,
		IsAlive:   s.IsAlive,
		Status:    s.GetStatus(),
		Equipment: EquipmentInfo{
			Weapon: s.Equipment,
			Armor:  "none",
			Ammo:   s.Ammo,
		},
		Inventory:      s.Inventory,
		ConnectedAt:    s.ConnectedAt.UnixMilli(),
		LastActivityAt: s.LastActivityAt.UnixMilli(),
		RemoteIP:       s.RemoteIP,
		Ping:           s.Ping,
		PacketLoss:     0,
		Statistics:     s.Statistics,
		IsMuted:        s.IsMuted,
		MuteEndTime:    s.MuteEndTime.UnixMilli(),
		JoinedAt:       s.ConnectedAt.UnixMilli(),
	}
}

// Info 玩家信息摘要（用于管理员列表）。
type Info struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Position       utils.Vec3 `json:"position"`
	Health         int        `json:"health"`
	Status         string     `json:"status"`
	ConnectedAt    int64      `json:"connectedAt"`
	LastActivityAt int64      `json:"lastActivityAt"`
	Ping           int64      `json:"ping"`
	Equipment      string     `json:"equipment"`
}

// Details 玩家详细信息（用于管理员详情弹窗）。
type Details struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Position       utils.Vec3      `json:"position"`
	Velocity       utils.Vec3      `json:"velocity"`
	Rotation       utils.Rotation  `json:"rotation"`
	Health         int             `json:"health"`
	MaxHealth      int             `json:"maxHealth"`
	IsAlive        bool            `json:"isAlive"`
	Status         string          `json:"status"`
	Equipment      EquipmentInfo   `json:"equipment"`
	Inventory      []InventoryItem `json:"inventory"`
	ConnectedAt    int64           `json:"connectedAt"`
	LastActivityAt int64           `json:"lastActivityAt"`
	RemoteIP       string          `json:"remoteIP"`
	Ping           int64           `json:"ping"`
	PacketLoss     float64         `json:"packetLoss"`
	Statistics     PlayerStats     `json:"statistics"`
	IsMuted        bool            `json:"isMuted"`
	MuteEndTime    int64           `json:"muteEndTime,omitempty"`
	JoinedAt       int64           `json:"joinedAt"`
}

// EquipmentInfo 装备信息。
type EquipmentInfo struct {
	Weapon string `json:"weapon"`
	Armor  string `json:"armor"`
	Ammo   int    `json:"ammo"`
}
