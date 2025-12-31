package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Node    NodeConfig    `json:"node"`
	P2P     P2PConfig     `json:"p2p"`
	Storage StorageConfig `json:"storage"`
	AI      AIConfig      `json:"ai"`
	Logging LoggingConfig `json:"logging"`
}

type NodeConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type P2PConfig struct {
	ListenPort      int      `json:"listen_port"`
	BootstrapPeers  []string `json:"bootstrap_peers"`
	MaxPeers        int      `json:"max_peers"`
	EnableDiscovery bool     `json:"enable_discovery"`
}

type StorageConfig struct {
	DataDir       string `json:"data_dir"`
	MaxSizeGB     int    `json:"max_size_gb"`
	EnableBackups bool   `json:"enable_backups"`
}

type AIConfig struct {
	Endpoint       string `json:"endpoint"`
	Timeout        int    `json:"timeout"`
	MaxRetries     int    `json:"max_retries"`
	EnableOffline  bool   `json:"enable_offline_queue"`
}

type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputFile string `json:"output_file"`
}

func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".synapse", "data")

	return &Config{
		Node: NodeConfig{
			ID:   "",
			Name: "synapse-node",
		},
		P2P: P2PConfig{
			ListenPort:      8080,
			BootstrapPeers:  []string{},
			MaxPeers:        50,
			EnableDiscovery: false,
		},
		Storage: StorageConfig{
			DataDir:       dataDir,
			MaxSizeGB:     10,
			EnableBackups: true,
		},
		AI: AIConfig{
			Endpoint:      "https://svceai.site/api/chat",
			Timeout:       30,
			MaxRetries:    3,
			EnableOffline: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputFile: "",
		},
	}
}

func Load(path string) (*Config, error) {
	if path == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	if c.P2P.ListenPort < 1024 || c.P2P.ListenPort > 65535 {
		return fmt.Errorf("invalid P2P listen port: %d", c.P2P.ListenPort)
	}

	if c.P2P.MaxPeers < 1 {
		return fmt.Errorf("max peers must be at least 1")
	}

	if c.Storage.MaxSizeGB < 1 {
		return fmt.Errorf("max storage size must be at least 1 GB")
	}

	if c.AI.Timeout < 1 {
		return fmt.Errorf("AI timeout must be at least 1 second")
	}

	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}
