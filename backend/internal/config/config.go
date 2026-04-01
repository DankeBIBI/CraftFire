// Package config 提供 CraftFire 应用的配置管理功能。
// 支持从 YAML 文件和环境变量加载配置。
package config

import (
	"os"
	"strconv"
)

// Config 应用配置结构体，包含所有子系统的配置项。
type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	LAN      LANConfig      `yaml:"lan"`
	Relay    RelayConfig    `yaml:"relay"`
	World    WorldConfig    `yaml:"world"`
	Physics  PhysicsConfig  `yaml:"physics"`
	Client   ClientConfig   `yaml:"client"`
	UI       UIConfig       `yaml:"ui"`
	Logging  LoggingConfig  `yaml:"logging"`
	Security SecurityConfig `yaml:"security"`
}

// AppConfig 应用基础信息。
type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// ServerConfig 服务端配置。
type ServerConfig struct {
	DefaultTickRate      int `yaml:"defaultTickRate"`
	MaxRoomsPerIP        int `yaml:"maxRoomsPerIP"`
	MaxConcurrentPlayers int `yaml:"maxConcurrentPlayers"`
	RoomIdleTimeout      int `yaml:"roomIdleTimeout"`
	PlayerIdleTimeout    int `yaml:"playerIdleTimeout"`
}

// LANConfig 局域网发现配置。
type LANConfig struct {
	BroadcastInterval int    `yaml:"broadcastInterval"`
	DiscoveryPort     int    `yaml:"discoveryPort"`
	BroadcastAddr     string `yaml:"broadcastAddr"`
}

// WorldConfig 世界配置。
type WorldConfig struct {
	ChunkSize           int  `yaml:"chunkSize"`
	RenderDistance      int  `yaml:"renderDistance"`
	WorldHeightLimit    int  `yaml:"worldHeightLimit"`
	PersistenceInterval int  `yaml:"persistenceInterval"`
	AutoSave            bool `yaml:"autoSave"`
}

// PhysicsConfig 物理引擎配置。
type PhysicsConfig struct {
	Gravity         float64 `yaml:"gravity"`
	DragCoefficient float64 `yaml:"dragCoefficient"`
	MaxVelocity     float64 `yaml:"maxVelocity"`
	Authority       string  `yaml:"authority"`
	UseWasm         bool    `yaml:"useWasm"`
	WorkerThreshold int     `yaml:"workerThreshold"`
}

// ClientConfig 客户端配置。
type ClientConfig struct {
	TargetFPS             int  `yaml:"targetFPS"`
	InterpolationDuration int  `yaml:"interpolationDuration"`
	PointerLockRequired   bool `yaml:"pointerLockRequired"`
	DefaultRenderDistance int  `yaml:"defaultRenderDistance"`
}

// UIConfig 用户界面配置。
type UIConfig struct {
	Language      string  `yaml:"language"`
	UIScale       float64 `yaml:"uiScale"`
	ShowDebugInfo bool    `yaml:"showDebugInfo"`
}

// LoggingConfig 日志配置。
type LoggingConfig struct {
	Level         string `yaml:"level"`
	Format        string `yaml:"format"`
	Output        string `yaml:"output"`
	FilePath      string `yaml:"filePath"`
	MaxLogSize    int    `yaml:"maxLogSize"`
	MaxLogBackups int    `yaml:"maxLogBackups"`
}

// SecurityConfig 安全配置。
type SecurityConfig struct {
	RoomPasswordRequired bool   `yaml:"roomPasswordRequired"`
	EnableTLS            bool   `yaml:"enableTLS"`
	RateLimitPerSecond   int    `yaml:"rateLimitPerSecond"`
	AdminPassword        string `yaml:"adminPassword"`
	AdminPasswordHash    string `yaml:"adminPasswordHash"`
	EnableAdminPanel     bool   `yaml:"enableAdminPanel"`
}

// RelayConfig 互联网中继服务器配置。
type RelayConfig struct {
	// 是否启用 relay 服务器
	Enabled bool `yaml:"enabled"`
	// 监听地址（relay 服务器模式专用）
	Addr string `yaml:"addr"`
	// TLS 证书（可选）
	TLSCert string `yaml:"tlsCert"`
	TLSKey  string `yaml:"tlsKey"`
}

// Load 从 YAML 文件加载配置，并用环境变量覆盖。
func Load(filePath string) (*Config, error) {
	cfg := LoadDefault()

	// 如果配置文件存在就加载它
	if _, err := os.Stat(filePath); err == nil {
		// TODO: 使用 yaml.v3 解析文件
		_ = filePath
	}

	// 从环境变量覆盖关键配置
	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides 从环境变量中读取值覆盖配置。
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		cfg.Security.AdminPassword = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("MAX_PLAYERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.MaxConcurrentPlayers = n
		}
	}
}
