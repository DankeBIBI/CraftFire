package model

import (
	applogger "CraftFire/backend/internal/logger"
)

// Syncer 局域网模型同步器，负责在房间内共享和同步 3D 模型。
type Syncer struct {
	manager *Manager
}

// NewSyncer 创建一个新的模型同步器。
func NewSyncer(manager *Manager) *Syncer {
	return &Syncer{manager: manager}
}

// DiscoverRemoteModels 发现局域网内其他设备上的可用模型。
// TODO: 通过 WebSocket 或 HTTP 查询远程设备的模型列表。
func (s *Syncer) DiscoverRemoteModels(remoteIP string) ([]Info, error) {
	applogger.Info("正在发现远程模型 (IP: %s)", remoteIP)
	// 存根实现
	return []Info{}, nil
}

// SyncModel 从远程设备下载指定模型到本地。
// 支持 MD5 校验和增量下载。
func (s *Syncer) SyncModel(modelId string, remoteIP string) error {
	applogger.Info("开始同步模型 %s (来源: %s)", modelId, remoteIP)
	// TODO: 实现断点续传下载
	return nil
}

// GetMissingModels 比对本地与远程模型列表，返回缺失的模型信息。
func (s *Syncer) GetMissingModels(remoteModels []Info) []Info {
	missing := make([]Info, 0)
	localModels, _ := s.manager.ListModels("")

	localMap := make(map[string]bool)
	for _, m := range localModels {
		localMap[m.MD5Hash] = true
	}

	for _, remote := range remoteModels {
		if !localMap[remote.MD5Hash] {
			missing = append(missing, remote)
		}
	}

	return missing
}
