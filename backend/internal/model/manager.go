// Package model 提供 3D 模型的导入、存储、验证和局域网同步功能。
package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"CraftFire/backend/internal/config"
	applogger "CraftFire/backend/internal/logger"
	"CraftFire/backend/internal/utils"
)

// Info 3D 模型信息。
type Info struct {
	ModelID      string         `json:"modelId"`
	Name         string         `json:"name"`
	Format       string         `json:"format"`
	FileSize     int64          `json:"fileSize"`
	FilePath     string         `json:"filePath"`
	MD5Hash      string         `json:"md5Hash"`
	UploadedAt   int64          `json:"uploadedAt"`
	UploadedBy   string         `json:"uploadedBy"`
	ThumbnailURL string         `json:"thumbnailUrl,omitempty"`
	Metadata     *ModelMetadata `json:"metadata,omitempty"`
	Version      int            `json:"version"`
	IsPublic     bool           `json:"isPublic"`
}

// ModelMetadata 模型元数据。
type ModelMetadata struct {
	VertexCount   int  `json:"vertexCount,omitempty"`
	TriangleCount int  `json:"triangleCount,omitempty"`
	Materials     int  `json:"materials,omitempty"`
	Textures      int  `json:"textures,omitempty"`
	HasAnimations bool `json:"hasAnimations,omitempty"`
}

// Manager 3D 模型管理器。
type Manager struct {
	cfg      *config.Config
	models   map[string]*Info
	mu       sync.RWMutex
	storeDir string
}

// NewManager 创建一个新的模型管理器。
func NewManager(cfg *config.Config) *Manager {
	homeDir, _ := os.UserHomeDir()
	storeDir := filepath.Join(homeDir, ".CraftFire", "models")

	// 确保存储目录存在
	os.MkdirAll(storeDir, 0755)

	return &Manager{
		cfg:      cfg,
		models:   make(map[string]*Info),
		storeDir: storeDir,
	}
}

// ImportModel 导入 3D 模型文件。
//
// 参数：
//   - filePath: 源文件路径
//   - roomId: 关联的房间号（可为空）
//
// 返回值：模型 ID
func (m *Manager) ImportModel(filePath string, roomId string) (string, error) {
	// 验证文件格式
	if err := ValidateFile(filePath); err != nil {
		return "", fmt.Errorf("模型文件验证失败: %w", err)
	}

	modelID := utils.GenerateUUID()

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件信息失败: %w", err)
	}

	// 计算 MD5
	hash, err := CalculateMD5(filePath)
	if err != nil {
		return "", fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 复制到存储目录
	destPath := filepath.Join(m.storeDir, modelID+filepath.Ext(filePath))
	if err := CopyFile(filePath, destPath); err != nil {
		return "", fmt.Errorf("复制模型文件失败: %w", err)
	}

	info := &Info{
		ModelID:  modelID,
		Name:     fileInfo.Name(),
		Format:   filepath.Ext(filePath)[1:],
		FileSize: fileInfo.Size(),
		FilePath: destPath,
		MD5Hash:  hash,
		Version:  1,
		IsPublic: true,
	}

	m.mu.Lock()
	m.models[modelID] = info
	m.mu.Unlock()

	applogger.Info("模型已导入: %s (%s)", info.Name, modelID)
	return modelID, nil
}

// ListModels 获取可用的模型列表。
func (m *Manager) ListModels(roomId string) ([]Info, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Info, 0, len(m.models))
	for _, info := range m.models {
		list = append(list, *info)
	}
	return list, nil
}

// GetModelInfo 获取指定模型的详细信息。
func (m *Manager) GetModelInfo(modelId string) (*Info, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.models[modelId]
	if !exists {
		return nil, fmt.Errorf("模型 %s 不存在", modelId)
	}
	return info, nil
}

// DeleteModel 删除已导入的模型。
func (m *Manager) DeleteModel(modelId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.models[modelId]
	if !exists {
		return fmt.Errorf("模型 %s 不存在", modelId)
	}

	// 删除文件
	os.Remove(info.FilePath)
	delete(m.models, modelId)

	applogger.Info("模型已删除: %s", modelId)
	return nil
}

// SyncModelsInLAN 同步局域网内可用的模型（存根实现）。
func (m *Manager) SyncModelsInLAN() ([]Info, error) {
	// TODO: 实现局域网模型发现与同步
	applogger.Info("开始局域网模型同步...")
	return []Info{}, nil
}

// DownloadFromLAN 从局域网其他设备下载模型（存根实现）。
func (m *Manager) DownloadFromLAN(modelId string, sourceIP string) error {
	// TODO: 实现从局域网设备下载模型
	applogger.Info("开始下载模型 %s (来源: %s)", modelId, sourceIP)
	return nil
}
