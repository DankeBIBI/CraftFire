package admin

import (
	"fmt"
	"time"

	"CraftFire/backend/internal/room"
)

// RoomStatistics 房间统计数据结构体。
type RoomStatistics struct {
	RoomID             string `json:"roomId"`
	TotalPlayers       int    `json:"totalPlayers"`
	MaxPlayers         int    `json:"maxPlayers"`
	TotalPlayersJoined int    `json:"totalPlayersJoined"`
	Uptime             int64  `json:"uptime"`
	TotalBlocksPlaced  int64  `json:"totalBlocksPlaced"`
	TotalBlocksRemoved int64  `json:"totalBlocksRemoved"`
	AveragePing        int64  `json:"averagePing"`
	PeakPlayerCount    int    `json:"peakPlayerCount"`
	CreatedAt          int64  `json:"createdAt"`
	LastUpdated        int64  `json:"lastUpdated"`
}

// Stats 房间统计数据收集服务。
type Stats struct {
	roomManager *room.Manager
}

// NewStats 创建一个新的房间统计服务实例。
func NewStats(rm *room.Manager) *Stats {
	return &Stats{roomManager: rm}
}

// GetRoomStatistics 获取指定房间的统计数据。
//
// 参数：
//   - roomId: 房间号
//
// 返回值：房间统计信息
func (s *Stats) GetRoomStatistics(roomId string) (*RoomStatistics, error) {
	r, err := s.roomManager.GetRoom(roomId)
	if err != nil {
		return nil, fmt.Errorf("获取房间 %s 统计数据失败: %w", roomId, err)
	}

	uptime := time.Since(r.CreatedAt).Seconds()

	return &RoomStatistics{
		RoomID:             r.ID,
		TotalPlayers:       r.GetPlayerCount(),
		MaxPlayers:         r.MaxPlayers,
		TotalPlayersJoined: r.TotalPlayersJoined,
		Uptime:             int64(uptime),
		TotalBlocksPlaced:  r.TotalBlocksPlaced,
		TotalBlocksRemoved: r.TotalBlocksRemoved,
		AveragePing:        0, // TODO: 计算平均延迟
		PeakPlayerCount:    r.PeakPlayerCount,
		CreatedAt:          r.CreatedAt.UnixMilli(),
		LastUpdated:        time.Now().UnixMilli(),
	}, nil
}
