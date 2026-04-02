package room

// RoomConfig 房间配置结构体。
type RoomConfig struct {
	RoomID         string `json:"roomId"`
	Port           int    `json:"port"`
	MaxPlayers     int    `json:"maxPlayers"`
	CurrentPlayers int    `json:"currentPlayers"`
	WorldSeed      string `json:"worldSeed"`
	CreatedAt      int64  `json:"createdAt"`
	LastActivityAt int64  `json:"lastActivityAt"`
	IsPublic       bool   `json:"isPublic"`
	GameMode       string `json:"gameMode"`
	IsLocked       bool   `json:"isLocked"`
}
