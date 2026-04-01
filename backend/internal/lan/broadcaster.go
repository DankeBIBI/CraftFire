package lan

import (
	"encoding/json"
	"fmt"
	"net"

	"CraftFire/backend/internal/config"
	applogger "CraftFire/backend/internal/logger"
)

// Broadcaster 负责向局域网广播房间信息。
type Broadcaster struct {
	cfg  *config.Config
	conn *net.UDPConn
}

// NewBroadcaster 创建一个新的广播器实例。
func NewBroadcaster(cfg *config.Config) *Broadcaster {
	return &Broadcaster{cfg: cfg}
}

// BroadcastOnce 发送一次 UDP 广播。
func (b *Broadcaster) BroadcastOnce(info ServerInfo) error {
	addr := fmt.Sprintf("%s:%d", b.cfg.LAN.BroadcastAddr, b.cfg.LAN.DiscoveryPort)
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return fmt.Errorf("解析广播地址失败: %w", err)
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("创建广播连接失败: %w", err)
	}
	defer conn.Close()

	data, err := json.Marshal([]ServerInfo{info})
	if err != nil {
		return fmt.Errorf("序列化广播数据失败: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送广播失败: %w", err)
	}

	applogger.Debug("已广播房间信息: %s", info.RoomID)
	return nil
}
