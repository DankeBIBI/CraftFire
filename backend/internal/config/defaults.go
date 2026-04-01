package config

// LoadDefault 返回一份带有默认值的配置实例。
// 所有默认值来源于 config.example.yml 中的推荐值。
func LoadDefault() *Config {
	return &Config{
		App: AppConfig{
			Name:        "CraftFire",
			Version:     "1.0.0",
			Description: "Sandbox + FPS Hybrid Game",
		},
		Server: ServerConfig{
			DefaultTickRate:      8,
			MaxRoomsPerIP:        5,
			MaxConcurrentPlayers: 200,
			RoomIdleTimeout:      3600,
			PlayerIdleTimeout:    300,
		},
		LAN: LANConfig{
			BroadcastInterval: 5,
			DiscoveryPort:     9999,
			BroadcastAddr:     "255.255.255.255",
		},
		Relay: RelayConfig{
			Enabled:   false,
			Addr:      ":8848",
			TLSCert:  "",
			TLSKey:   "",
		},
		World: WorldConfig{
			ChunkSize:           16,
			RenderDistance:      8,
			WorldHeightLimit:    256,
			PersistenceInterval: 300,
			AutoSave:            true,
		},
		Physics: PhysicsConfig{
			Gravity:         9.8,
			DragCoefficient: 0.1,
			MaxVelocity:     100,
			Authority:       "server",
			UseWasm:         false,
			WorkerThreshold: 1000,
		},
		Client: ClientConfig{
			TargetFPS:             60,
			InterpolationDuration: 125,
			PointerLockRequired:   true,
			DefaultRenderDistance: 6,
		},
		UI: UIConfig{
			Language:      "zh-CN",
			UIScale:       1.0,
			ShowDebugInfo: false,
		},
		Logging: LoggingConfig{
			Level:         "INFO",
			Format:        "text",
			Output:        "console",
			FilePath:      "./logs/craftfire.log",
			MaxLogSize:    10485760,
			MaxLogBackups: 5,
		},
		Security: SecurityConfig{
			RoomPasswordRequired: false,
			EnableTLS:            false,
			RateLimitPerSecond:   100,
			AdminPassword:        "admin123",
			AdminPasswordHash:    "",
			EnableAdminPanel:     true,
		},
	}
}
