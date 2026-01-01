package p2p

import (
	"context"
	"testing"

	"github.com/princetheprogrammer/synapse/internal/config"
	"github.com/princetheprogrammer/synapse/internal/logger"
	"github.com/princetheprogrammer/synapse/pkg/p2p/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedNetworkInitialization(t *testing.T) {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	network, err := New(cfg, log, "test-node-id")
	require.NoError(t, err)

	assert.NotNil(t, network.encryptor)
	assert.NotNil(t, network.handshakeMgr)
	assert.NotNil(t, network.bootstrapMgr)
	assert.NotNil(t, network.topologyMgr)
	assert.NotNil(t, network.monitor)
	assert.NotNil(t, network.peerExchange)
}

func TestNetworkReport(t *testing.T) {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	network, err := New(cfg, log, "test-node-id")
	require.NoError(t, err)

	// Start the network
	err = network.Start(ctx)
	require.NoError(t, err)

	// Get network report
	report := network.GetNetworkReport()
	assert.NotNil(t, report)

	// Check that report contains expected keys
	assert.Contains(t, report, "stats")
	assert.Contains(t, report, "peer_qualities")
	assert.Contains(t, report, "unhealthy_peers")
	assert.Contains(t, report, "bandwidth")
	assert.Contains(t, report, "topology_metrics")

	// Stop the network
	err = network.Stop()
	assert.NoError(t, err)
}

func TestTopologyMetrics(t *testing.T) {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	network, err := New(cfg, log, "test-node-id")
	require.NoError(t, err)

	metrics := network.GetTopologyMetrics()
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "total_peers")
	assert.Contains(t, metrics, "connected_peers")
	assert.Contains(t, metrics, "topology_type")
}

func TestBootstrapManager(t *testing.T) {
	nodes := []string{"192.168.1.1:8080", "192.168.1.2:8080"}
	manager := discovery.NewBootstrapManager(nodes)

	// Test initial nodes
	assert.Equal(t, nodes, manager.GetNodes())

	// Test adding a new node
	manager.AddNode("192.168.1.3:8080")
	updatedNodes := manager.GetNodes()
	assert.Len(t, updatedNodes, 3)
	assert.Contains(t, updatedNodes, "192.168.1.3:8080")
}

func TestConnectionQuality(t *testing.T) {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	network, err := New(cfg, log, "test-node-id")
	require.NoError(t, err)

	// Initially should not have quality metrics for any peer
	_, exists := network.GetConnectionQuality("nonexistent-peer")
	assert.False(t, exists)
}