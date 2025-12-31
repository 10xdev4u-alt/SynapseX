package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.NotNil(t, cfg)
	assert.Equal(t, "synapse-node", cfg.Node.Name)
	assert.Equal(t, 8080, cfg.P2P.ListenPort)
	assert.Equal(t, "https://svceai.site/api/chat", cfg.AI.Endpoint)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*Config)
		expectErr bool
	}{
		{
			name:      "valid default config",
			modify:    func(c *Config) {},
			expectErr: false,
		},
		{
			name: "invalid port too low",
			modify: func(c *Config) {
				c.P2P.ListenPort = 80
			},
			expectErr: true,
		},
		{
			name: "invalid port too high",
			modify: func(c *Config) {
				c.P2P.ListenPort = 70000
			},
			expectErr: true,
		},
		{
			name: "invalid max peers",
			modify: func(c *Config) {
				c.P2P.MaxPeers = 0
			},
			expectErr: true,
		},
		{
			name: "invalid storage size",
			modify: func(c *Config) {
				c.Storage.MaxSizeGB = 0
			},
			expectErr: true,
		},
		{
			name: "invalid log level",
			modify: func(c *Config) {
				c.Logging.Level = "invalid"
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(cfg)
			err := cfg.Validate()

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	originalCfg := Default()
	originalCfg.Node.Name = "test-node"
	originalCfg.P2P.ListenPort = 9090

	err := originalCfg.Save(configPath)
	require.NoError(t, err)

	assert.FileExists(t, configPath)

	loadedCfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "test-node", loadedCfg.Node.Name)
	assert.Equal(t, 9090, loadedCfg.P2P.ListenPort)
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/non/existent/path.json")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, Default().Node.Name, cfg.Node.Name)
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(configPath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	_, err = Load(configPath)
	assert.Error(t, err)
}
