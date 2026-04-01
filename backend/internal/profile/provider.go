// Package profile 提供玩家个人信息管理功能。
package profile

import (
	"fmt"
	"time"

	"CraftFire/backend/internal/player"
)

// PlayerProfile 玩家完整个人资料。
type PlayerProfile struct {
	PlayerID       string                 `json:"playerId"`
	Nickname       string                 `json:"nickname"`
	Avatar         string                 `json:"avatar,omitempty"`
	CharacterModel string                 `json:"characterModel"`
	JoinedAt       int64                  `json:"joinedAt"`
	LastSeenAt     int64                  `json:"lastSeenAt"`
	TotalPlayTime  int64                  `json:"totalPlayTime"`
	Level          int                    `json:"level"`
	Experience     int                    `json:"experience"`
	NextLevelExp   int                    `json:"nextLevelExp"`
	Customization  Customization          `json:"customization"`
	Equipment      EquipmentData          `json:"equipment"`
	Inventory      []player.InventoryItem `json:"inventory"`
}

// Customization 角色定制数据。
type Customization struct {
	SkinColor     string   `json:"skinColor,omitempty"`
	ClothingStyle string   `json:"clothingStyle,omitempty"`
	Accessories   []string `json:"accessories,omitempty"`
}

// EquipmentData 装备数据。
type EquipmentData struct {
	Weapon string `json:"weapon,omitempty"`
	Armor  string `json:"armor,omitempty"`
	Ammo   int    `json:"ammo"`
}

// PlayerStatistics 玩家统计数据。
type PlayerStatistics struct {
	PlayerID           string   `json:"playerId"`
	TotalBlocksPlaced  int64    `json:"totalBlocksPlaced"`
	TotalBlocksRemoved int64    `json:"totalBlocksRemoved"`
	TotalKills         int      `json:"totalKills"`
	TotalDeaths        int      `json:"totalDeaths"`
	DistanceTraveled   float64  `json:"distanceTraveled"`
	GameTime           int64    `json:"gameTime"`
	RoomsVisited       int      `json:"roomsVisited"`
	Achievements       []string `json:"achievements"`
	LastUpdated        int64    `json:"lastUpdated"`
}

// Provider 玩家资料数据提供者。
type Provider struct {
	playerManager *player.Manager
}

// NewProvider 创建一个新的玩家资料提供者。
func NewProvider(pm *player.Manager) *Provider {
	return &Provider{playerManager: pm}
}

// GetProfile 获取玩家资料。
func (p *Provider) GetProfile(playerId string) (*PlayerProfile, error) {
	state, err := p.playerManager.GetPlayer(playerId)
	if err != nil {
		return nil, fmt.Errorf("获取玩家资料失败: %w", err)
	}

	return &PlayerProfile{
		PlayerID:       state.ID,
		Nickname:       state.Name,
		CharacterModel: "default_player.glb",
		JoinedAt:       state.ConnectedAt.UnixMilli(),
		LastSeenAt:     time.Now().UnixMilli(),
		TotalPlayTime:  int64(time.Since(state.ConnectedAt).Seconds()),
		Level:          1,
		Experience:     0,
		NextLevelExp:   1000,
		Customization:  Customization{},
		Equipment: EquipmentData{
			Weapon: state.Equipment,
			Ammo:   state.Ammo,
		},
		Inventory: state.Inventory,
	}, nil
}

// UpdateProfile 更新玩家个人信息。
func (p *Provider) UpdateProfile(nickname string, skinColor string) error {
	// TODO: 实现个人信息更新
	return nil
}

// GetStatistics 获取玩家统计数据。
func (p *Provider) GetStatistics(playerId string) (*PlayerStatistics, error) {
	state, err := p.playerManager.GetPlayer(playerId)
	if err != nil {
		return nil, fmt.Errorf("获取玩家统计失败: %w", err)
	}

	return &PlayerStatistics{
		PlayerID:           state.ID,
		TotalBlocksPlaced:  state.Statistics.BlocksPlaced,
		TotalBlocksRemoved: state.Statistics.BlocksRemoved,
		TotalKills:         state.Statistics.KillCount,
		TotalDeaths:        state.Statistics.DeathCount,
		DistanceTraveled:   state.Statistics.DistanceTraveled,
		GameTime:           int64(time.Since(state.ConnectedAt).Seconds()),
		Achievements:       []string{},
		LastUpdated:        time.Now().UnixMilli(),
	}, nil
}
