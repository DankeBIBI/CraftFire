// Package admin 提供管理员面板的身份验证与授权功能。
package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"CraftFire/backend/internal/config"
	applogger "CraftFire/backend/internal/logger"
)

// sessionDuration 会话令牌有效期（30 分钟）
const sessionDuration = 30 * time.Minute

// SessionInfo 管理员会话信息。
type SessionInfo struct {
	Token     string
	ExpiresAt time.Time
}

// Authentication 管理员身份验证服务。
type Authentication struct {
	cfg      *config.Config
	sessions map[string]*SessionInfo // roomId -> SessionInfo
	mu       sync.RWMutex
}

// NewAuthentication 创建一个新的管理员验证服务实例。
func NewAuthentication(cfg *config.Config) *Authentication {
	return &Authentication{
		cfg:      cfg,
		sessions: make(map[string]*SessionInfo),
	}
}

// Verify 验证管理员密码。
// 验证成功后生成短期会话令牌。
//
// 参数：
//   - roomId: 房间号
//   - password: 管理员密码
//
// 返回值：
//   - token: 会话令牌
//   - expiresAt: 过期时间戳（毫秒）
//   - error: 验证失败时返回错误
func (a *Authentication) Verify(roomId string, password string) (string, int64, error) {
	if !a.cfg.Security.EnableAdminPanel {
		return "", 0, fmt.Errorf("管理员面板未启用")
	}

	// 验证密码
	if password != a.cfg.Security.AdminPassword {
		applogger.Warn("管理员验证失败 (房间 %s): 密码错误", roomId)
		return "", 0, fmt.Errorf("密码错误")
	}

	// 生成会话令牌
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(sessionDuration)

	a.mu.Lock()
	a.sessions[roomId] = &SessionInfo{
		Token:     token,
		ExpiresAt: expiresAt,
	}
	a.mu.Unlock()

	applogger.Info("管理员已验证 (房间 %s)", roomId)
	return token, expiresAt.UnixMilli(), nil
}

// ValidateToken 校验会话令牌是否有效。
func (a *Authentication) ValidateToken(roomId string, token string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	session, exists := a.sessions[roomId]
	if !exists {
		return false
	}

	if session.Token != token {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(a.sessions, roomId)
		return false
	}

	return true
}

// Logout 注销管理员会话。
func (a *Authentication) Logout(roomId string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, roomId)
	applogger.Info("管理员已注销 (房间 %s)", roomId)
}
