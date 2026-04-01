// Package lan 提供局域网服务发现功能。
// 使用 UDP 广播来搜索局域网内运行 CraftFire 的其他设备。
package lan

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"CraftFire/backend/internal/config"
	applogger "CraftFire/backend/internal/logger"
)

// ServerInfo 局域网发现的服务器信息。
type ServerInfo struct {
	RoomID      string `json:"roomId"`
	IP          string `json:"ip"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	GameMode    string `json:"gameMode"`
}

// Discovery 局域网发现服务。
type Discovery struct {
	cfg        *config.Config
	conn       *net.UDPConn
	servers    []ServerInfo
	mu         sync.RWMutex
	stopCh     chan struct{}
	isRunning  bool
	localRooms []ServerInfo // 本地运行的房间信息（用于广播）
}

// NewDiscovery 创建一个新的局域网发现服务实例。
func NewDiscovery(cfg *config.Config) *Discovery {
	return &Discovery{
		cfg:        cfg,
		servers:    make([]ServerInfo, 0),
		stopCh:     make(chan struct{}),
		localRooms: make([]ServerInfo, 0),
	}
}

// StartBroadcast 启动 UDP 广播，定期向局域网内广播本地房间信息。
// 应在独立的 goroutine 中运行。
func (d *Discovery) StartBroadcast() {
	addr := fmt.Sprintf("%s:%d", d.cfg.LAN.BroadcastAddr, d.cfg.LAN.DiscoveryPort)
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		applogger.Error("解析 UDP 广播地址失败: %v", err)
		return
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		applogger.Error("创建 UDP 广播连接失败: %v", err)
		return
	}
	defer conn.Close()

	d.isRunning = true
	applogger.Info("局域网广播服务已启动 (端口 %d)", d.cfg.LAN.DiscoveryPort)

	ticker := time.NewTicker(time.Duration(d.cfg.LAN.BroadcastInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.mu.RLock()
			if len(d.localRooms) > 0 {
				data, _ := json.Marshal(d.localRooms)
				_, err := conn.Write(data)
				if err != nil {
					applogger.Debug("UDP 广播发送失败: %v", err)
				}
			}
			d.mu.RUnlock()

		case <-d.stopCh:
			d.isRunning = false
			applogger.Info("局域网广播服务已停止")
			return
		}
	}
}

// FindServers 搜索局域网内的可用游戏房间。
// 监听 UDP 广播并收集服务器信息。
//
// 返回值：发现的服务器列表
func (d *Discovery) FindServers() ([]ServerInfo, error) {
	addr := fmt.Sprintf(":%d", d.cfg.LAN.DiscoveryPort)
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("解析 UDP 监听地址失败: %w", err)
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("创建 UDP 监听失败: %w", err)
	}
	defer conn.Close()

	// 设置读取超时为 3 秒
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	servers := make([]ServerInfo, 0)
	buf := make([]byte, 4096)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			// 超时即正常结束搜索
			break
		}

		var roomInfos []ServerInfo
		if err := json.Unmarshal(buf[:n], &roomInfos); err != nil {
			continue
		}

		for i := range roomInfos {
			roomInfos[i].IP = remoteAddr.IP.String()
		}
		servers = append(servers, roomInfos...)
	}

	d.mu.Lock()
	d.servers = servers
	d.mu.Unlock()

	applogger.Info("局域网搜索完成，发现 %d 个房间", len(servers))
	return servers, nil
}

// RegisterRoom 注册本地房间到广播列表。
func (d *Discovery) RegisterRoom(info ServerInfo) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.localRooms = append(d.localRooms, info)
}

// UnregisterRoom 从广播列表移除本地房间。
func (d *Discovery) UnregisterRoom(roomID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, r := range d.localRooms {
		if r.RoomID == roomID {
			d.localRooms = append(d.localRooms[:i], d.localRooms[i+1:]...)
			break
		}
	}
}

// Stop 停止局域网发现服务。
func (d *Discovery) Stop() {
	if d.isRunning {
		close(d.stopCh)
	}
}
