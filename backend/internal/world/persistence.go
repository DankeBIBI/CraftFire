package world

import (
	"encoding/json"
	"os"
	"path/filepath"

	applogger "CraftFire/backend/internal/logger"
)

// Persistence 世界数据持久化管理器。
type Persistence struct {
	saveDir string
}

// NewPersistence 创建一个新的持久化管理器。
func NewPersistence(saveDir string) *Persistence {
	os.MkdirAll(saveDir, 0755)
	return &Persistence{saveDir: saveDir}
}

// SaveWorldData 将世界分块数据保存到文件。
type WorldSaveData struct {
	Blocks []Block `json:"blocks"`
	RoomID string  `json:"roomId"`
}

// Save 保存世界数据到 JSON 文件。
func (p *Persistence) Save(roomID string, blocks []Block) error {
	data := WorldSaveData{
		Blocks: blocks,
		RoomID: roomID,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(p.saveDir, roomID+".json")
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	applogger.Info("世界数据已保存: %s", filePath)
	return nil
}

// Load 从文件加载世界数据。
func (p *Persistence) Load(roomID string) ([]Block, error) {
	filePath := filepath.Join(p.saveDir, roomID+".json")

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			applogger.Info("未找到保存文件 %s，将创建新世界", filePath)
			return []Block{}, nil
		}
		return nil, err
	}

	var data WorldSaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	applogger.Info("世界数据已加载: %s (%d 个方块)", filePath, len(data.Blocks))
	return data.Blocks, nil
}
