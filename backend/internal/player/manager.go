package player

import (
	"fmt"
	"sync"

	applogger "CraftFire/backend/internal/logger"
	"CraftFire/backend/internal/utils"
)

// Manager 玩家管理器，负责所有玩家状态的增删改查。
type Manager struct {
	players map[string]*State
	mu      sync.RWMutex
}

// NewManager 创建一个新的玩家管理器。
func NewManager() *Manager {
	return &Manager{
		players: make(map[string]*State),
	}
}

// AddPlayer 添加一个新玩家。
func (m *Manager) AddPlayer(id, name string) *State {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := NewState(id, name)
	m.players[id] = state
	applogger.Info("玩家已添加: %s (%s)", name, id)
	return state
}

// RemovePlayer 移除指定玩家。
func (m *Manager) RemovePlayer(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, exists := m.players[id]; exists {
		applogger.Info("玩家已移除: %s (%s)", p.Name, id)
		delete(m.players, id)
	}
}

// GetPlayer 获取指定玩家的状态。
func (m *Manager) GetPlayer(id string) (*State, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.players[id]
	if !exists {
		return nil, fmt.Errorf("玩家 %s 不存在", id)
	}
	return p, nil
}

// UpdatePosition 更新玩家位置。
func (m *Manager) UpdatePosition(id string, pos utils.Vec3, rot utils.Rotation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.players[id]
	if !exists {
		return fmt.Errorf("玩家 %s 不存在", id)
	}

	p.Position = pos
	p.Rotation = rot
	return nil
}

// GetAllPlayers 获取所有玩家列表。
func (m *Manager) GetAllPlayers() []*State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*State, 0, len(m.players))
	for _, p := range m.players {
		list = append(list, p)
	}
	return list
}

// GetPlayerCount 返回当前玩家总数。
func (m *Manager) GetPlayerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.players)
}
