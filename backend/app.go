package main

import (
	"context"
	"fmt"
	"net"

	"CraftFire/backend/internal/admin"
	"CraftFire/backend/internal/config"
	"CraftFire/backend/internal/lan"
	applogger "CraftFire/backend/internal/logger"
	"CraftFire/backend/internal/model"
	"CraftFire/backend/internal/player"
	"CraftFire/backend/internal/profile"
	"CraftFire/backend/internal/relay"
	"CraftFire/backend/internal/room"
	"CraftFire/backend/internal/websocket"
)

// App 是 CraftFire 的核心应用结构体，封装了所有后端服务。
// 通过 Wails 绑定到前端，前端可以直接调用其公开方法。
type App struct {
	ctx           context.Context
	cfg           *config.Config
	roomManager   *room.Manager
	playerManager *player.Manager
	wsHub         *websocket.Hub
	lanDiscovery  *lan.Discovery
	adminAuth     *admin.Authentication
	adminStats    *admin.Stats
	modelManager  *model.Manager
	profileProv   *profile.Provider
	relayServer   *relay.Server // 互联网中继服务器（可选）
}

// NewApp 创建一个新的 App 实例并初始化所有子系统。
func NewApp() *App {
	cfg := config.LoadDefault()

	hub := websocket.NewHub()
	rm := room.NewManager(cfg, hub)
	pm := player.NewManager()
	discovery := lan.NewDiscovery(cfg)
	auth := admin.NewAuthentication(cfg)
	stats := admin.NewStats(rm)
	mm := model.NewManager(cfg)
	pp := profile.NewProvider(pm)

	return &App{
		cfg:           cfg,
		roomManager:   rm,
		playerManager: pm,
		wsHub:         hub,
		lanDiscovery:  discovery,
		adminAuth:     auth,
		adminStats:    stats,
		modelManager:  mm,
		profileProv:   pp,
		relayServer:   nil,
	}
}

// startup 在 Wails 应用启动时被调用，用于保存上下文和初始化服务。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	applogger.Info("CraftFire 后端服务已启动")

	// 启动 WebSocket Hub
	go a.wsHub.Run()

	// 启动局域网发现广播
	go a.lanDiscovery.StartBroadcast()

	// 启动互联网中继服务器（如已启用）
	if a.cfg.Relay.Enabled {
		a.relayServer = relay.NewServer(a.cfg.Relay.Addr)
		go func() {
			applogger.Info("[Relay] 互联网中继服务器已启用，监听 %s", a.cfg.Relay.Addr)
			if err := a.relayServer.Start(); err != nil {
				applogger.Error("[Relay] 服务器启动失败: %v", err)
			}
		}()
	}
}

// domReady 在前端 DOM 准备就绪时被调用。
func (a *App) domReady(ctx context.Context) {
	applogger.Info("前端 DOM 已就绪")
}

// shutdown 在应用关闭前被调用，用于清理资源和持久化数据。
func (a *App) shutdown(ctx context.Context) {
	applogger.Info("CraftFire 正在关闭，保存数据...")
	a.roomManager.ShutdownAll()
	a.lanDiscovery.Stop()
	if a.relayServer != nil {
		a.relayServer.Stop()
	}
	applogger.Info("CraftFire 已安全关闭")
}

// ============================================================
// 房间操作方法（暴露给前端）
// ============================================================

// CreateRoom 创建一个新的游戏房间。
// 生成 6 位随机房间号，启动对应端口的 WebSocket 服务。
// 返回房间号和可能的错误。
func (a *App) CreateRoom() (string, error) {
	roomID, err := a.roomManager.CreateRoom()
	if err != nil {
		return "", fmt.Errorf("创建房间失败: %w", err)
	}

	ip := "127.0.0.1"
	if localIP := detectLocalIPv4(); localIP != "" {
		ip = localIP
	}
	a.lanDiscovery.RegisterRoom(lan.ServerInfo{
		RoomID:      roomID,
		IP:          ip,
		PlayerCount: 1,
		MaxPlayers:  10,
		GameMode:    "sandbox",
	})

	applogger.Info("房间已创建: %s", roomID)
	return roomID, nil
}

// JoinRoom 加入指定房间。
// 参数：
//   - roomId: 6 位数字房间号
//   - ip: 房主 IP（本地加入时可传 127.0.0.1）
//
// 返回是否成功以及可能的错误。
func (a *App) JoinRoom(roomId string, ip string) (bool, error) {
	err := a.roomManager.JoinRoom(roomId, ip)
	if err != nil {
		return false, fmt.Errorf("加入房间失败: %w", err)
	}
	applogger.Info("已加入房间 %s (%s)", roomId, ip)
	return true, nil
}

// LeaveRoom 离开当前房间。
func (a *App) LeaveRoom() error {
	if currentRoom := a.roomManager.GetCurrentRoomID(); currentRoom != "" {
		a.lanDiscovery.UnregisterRoom(currentRoom)
	}
	return a.roomManager.LeaveCurrentRoom()
}

// detectLocalIPv4 检测当前设备的局域网 IPv4 地址。
func detectLocalIPv4() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			if v4 := ipNet.IP.To4(); v4 != nil {
				return v4.String()
			}
		}
	}

	return ""
}

// FindLANServers 发现局域网内的可用游戏房间。
// 返回房间列表（包含房间号、IP 和在线人数）。
func (a *App) FindLANServers() ([]lan.ServerInfo, error) {
	servers, err := a.lanDiscovery.FindServers()
	if err != nil {
		return nil, fmt.Errorf("局域网搜索失败: %w", err)
	}
	return servers, nil
}

// ============================================================
// 方块操作方法
// ============================================================

// PlaceBlock 在指定坐标放置方块。
// 参数：
//   - x, y, z: 体素坐标
//   - blockType: 方块类型 ("stone"|"wood"|"glass"|"dirt")
func (a *App) PlaceBlock(x, y, z int, blockType string) error {
	return a.roomManager.PlaceBlock(x, y, z, blockType)
}

// RemoveBlock 移除指定坐标的方块。
func (a *App) RemoveBlock(x, y, z int) error {
	return a.roomManager.RemoveBlock(x, y, z)
}

// ============================================================
// 管理员操作方法
// ============================================================

// VerifyAdminPassword 验证管理员密码。
// 返回是否合法以及会话令牌。
func (a *App) VerifyAdminPassword(roomId string, password string) (map[string]interface{}, error) {
	token, expiresAt, err := a.adminAuth.Verify(roomId, password)
	if err != nil {
		return nil, fmt.Errorf("管理员验证失败: %w", err)
	}
	return map[string]interface{}{
		"token":     token,
		"expiresAt": expiresAt,
	}, nil
}

// GetOnlinePlayers 获取房间内所有在线玩家列表。
func (a *App) GetOnlinePlayers(roomId string) ([]player.Info, error) {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return nil, err
	}
	return r.GetPlayerInfoList(), nil
}

// GetPlayerDetails 获取指定玩家的详细信息。
func (a *App) GetPlayerDetails(roomId string, playerId string) (*player.Details, error) {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return nil, err
	}
	return r.GetPlayerDetails(playerId)
}

// KickPlayer 踢除指定玩家。
func (a *App) KickPlayer(roomId string, playerId string, reason string) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	return r.KickPlayer(playerId, reason)
}

// MutePlayer 禁言指定玩家。
func (a *App) MutePlayer(roomId string, playerId string, durationSeconds int) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	return r.MutePlayer(playerId, durationSeconds)
}

// GetRoomStats 获取房间统计数据。
func (a *App) GetRoomStats(roomId string) (*admin.RoomStatistics, error) {
	return a.adminStats.GetRoomStatistics(roomId)
}

// SetRoomLocked 设置房间锁定状态。
func (a *App) SetRoomLocked(roomId string, locked bool) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	r.SetLocked(locked)
	return nil
}

// IsRoomLocked 获取房间锁定状态。
func (a *App) IsRoomLocked(roomId string) (bool, error) {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return false, err
	}
	return r.IsLocked, nil
}

// BroadcastToRoom 向房间广播公告消息。
func (a *App) BroadcastToRoom(roomId string, message string) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	return r.Broadcast(message)
}

// ChangeGameMode 切换游戏模式。
func (a *App) ChangeGameMode(roomId string, mode string) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	return r.UpdateConfig(0, mode, "", r.IsPublic)
}

// HealPlayer 为玩家恢复生命值。
func (a *App) HealPlayer(roomId string, playerId string) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	return r.HealPlayer(playerId)
}

// TeleportPlayer 传送玩家到指定位置。
func (a *App) TeleportPlayer(roomId string, playerId string, x, y, z float64) error {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return err
	}
	return r.TeleportPlayer(playerId, x, y, z)
}

// GetRoomConfig 获取房间配置。
func (a *App) GetRoomConfig(roomId string) (*room.RoomConfig, error) {
	r, err := a.roomManager.GetRoom(roomId)
	if err != nil {
		return nil, err
	}
	return &room.RoomConfig{
		RoomID:     r.ID,
		Port:       r.Port,
		MaxPlayers: r.MaxPlayers,
		GameMode:   r.GameMode,
		IsPublic:   r.IsPublic,
		IsLocked:   r.IsLocked,
		WorldSeed:  r.WorldSeed,
	}, nil
}

// ============================================================
// 3D 模型管理方法
// ============================================================

// ImportModel 导入 3D 模型文件。
func (a *App) ImportModel(filePath string, roomId string) (string, error) {
	modelId, err := a.modelManager.ImportModel(filePath, roomId)
	if err != nil {
		return "", fmt.Errorf("导入模型失败: %w", err)
	}
	return modelId, nil
}

// ListAvailableModels 获取可用的 3D 模型列表。
func (a *App) ListAvailableModels(roomId string) ([]model.Info, error) {
	return a.modelManager.ListModels(roomId)
}

// GetModelInfo 获取指定模型的详细信息。
func (a *App) GetModelInfo(modelId string) (*model.Info, error) {
	return a.modelManager.GetModelInfo(modelId)
}

// DeleteModel 删除已导入的 3D 模型。
func (a *App) DeleteModel(modelId string) error {
	return a.modelManager.DeleteModel(modelId)
}

// SyncModelsInLAN 同步局域网内所有可用的模型。
func (a *App) SyncModelsInLAN() ([]model.Info, error) {
	return a.modelManager.SyncModelsInLAN()
}

// DownloadModelFromLAN 从局域网其他设备下载指定模型。
func (a *App) DownloadModelFromLAN(modelId string, sourceIP string) error {
	return a.modelManager.DownloadFromLAN(modelId, sourceIP)
}

// ============================================================
// 玩家资料方法
// ============================================================

// GetPlayerProfile 获取玩家详细信息。
func (a *App) GetPlayerProfile(playerId string) (*profile.PlayerProfile, error) {
	return a.profileProv.GetProfile(playerId)
}

// UpdatePlayerProfile 更新玩家个人信息。
func (a *App) UpdatePlayerProfile(nickname string, skinColor string) error {
	return a.profileProv.UpdateProfile(nickname, skinColor)
}

// GetPlayerStatistics 获取玩家游戏统计数据。
func (a *App) GetPlayerStatistics(playerId string) (*profile.PlayerStatistics, error) {
	return a.profileProv.GetStatistics(playerId)
}
