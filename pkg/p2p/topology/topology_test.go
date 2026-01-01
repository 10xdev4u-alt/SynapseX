package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologyManager(t *testing.T) {
	manager := NewManager(10)

	// Test basic creation
	assert.Equal(t, 10, manager.maxPeers)
	assert.Equal(t, 10, manager.meshThreshold)

	// Test adding a peer
	peer := Peer{
		ID:       "test-peer",
		Address:  "127.0.0.1:8080",
		Version:  "1.0.0",
		LastSeen: time.Now(),
	}
	manager.AddPeer(peer)

	// Verify peer was added
	info, exists := manager.GetPeerInfo("test-peer")
	assert.True(t, exists)
	assert.Equal(t, "test-peer", info.ID)
	assert.Equal(t, "127.0.0.1:8080", info.Address)
}

func TestConnectionQuality(t *testing.T) {
	manager := NewManager(10)

	// Test quality score calculation
	quality := ConnectionQuality{
		Latency:    100 * time.Millisecond,
		Bandwidth:  10.0,
		PacketLoss: 1.0,
		Jitter:     5 * time.Millisecond,
	}
	
	score := manager.calculateQualityScore(quality)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestGetBestPeers(t *testing.T) {
	manager := NewManager(10)

	// Add multiple peers with different qualities
	peer1 := Peer{ID: "peer1", Address: "127.0.0.1:8081"}
	peer2 := Peer{ID: "peer2", Address: "127.0.0.1:8082"}
	
	manager.AddPeer(peer1)
	manager.AddPeer(peer2)

	// Update quality metrics
	quality1 := ConnectionQuality{
		Latency:    50 * time.Millisecond,
		Bandwidth:  20.0,
		PacketLoss: 0.1,
	}
	manager.UpdatePeerQuality("peer1", quality1)

	quality2 := ConnectionQuality{
		Latency:    200 * time.Millisecond,
		Bandwidth:  5.0,
		PacketLoss: 5.0,
	}
	manager.UpdatePeerQuality("peer2", quality2)

	// peer1 should be better than peer2 based on quality
	bestPeers := manager.GetBestPeers(2)
	require.Len(t, bestPeers, 2)
	assert.Equal(t, "peer1", bestPeers[0]) // Better quality should be first
}

func TestTopologyType(t *testing.T) {
	manager := NewManager(10)

	// Test star topology (small network)
	assert.Equal(t, "star", manager.GetTopologyType())

	// Add a few peers to make it a medium network
	for i := 0; i < 5; i++ {
		peer := Peer{
			ID:       "peer" + string(rune('0'+i)),
			Address:  "127.0.0.1:808" + string(rune('0'+i)),
		}
		manager.AddPeer(peer)
	}
	
	// Should now be full-mesh
	assert.Equal(t, "full-mesh", manager.GetTopologyType())

	// Add more peers to make it a large network
	for i := 5; i < 15; i++ {
		peer := Peer{
			ID:       "peer" + string(rune('0'+i)),
			Address:  "127.0.0.1:808" + string(rune('0'+i)),
		}
		manager.AddPeer(peer)
	}

	// Should now be partial-mesh
	assert.Equal(t, "partial-mesh", manager.GetTopologyType())
}

func TestReputationSystem(t *testing.T) {
	manager := NewManager(10)

	peer := Peer{ID: "test-peer", Address: "127.0.0.1:8080"}
	manager.AddPeer(peer)

	// Test reputation update
	manager.UpdatePeerReputation("test-peer", 0.8)
	info, exists := manager.GetPeerInfo("test-peer")
	assert.True(t, exists)
	assert.Equal(t, 0.8, info.Reputation)

	manager.UpdatePeerReputation("test-peer", -0.5)
	info, exists = manager.GetPeerInfo("test-peer")
	assert.True(t, exists)
	assert.Equal(t, -0.5, info.Reputation)
}

func TestNetworkMetrics(t *testing.T) {
	manager := NewManager(10)

	peer := Peer{ID: "test-peer", Address: "127.0.0.1:8080"}
	manager.AddPeer(peer)

	metrics := manager.GetNetworkMetrics()
	assert.Equal(t, 1, metrics["total_peers"])
	assert.Equal(t, 0, metrics["connected_peers"]) // Peer is not marked as connected by default
	assert.Equal(t, "star", metrics["topology_type"])
}

func TestGetOptimalPeersForBroadcast(t *testing.T) {
	manager := NewManager(10)

	// Add peers
	for i := 0; i < 5; i++ {
		peer := Peer{
			ID:       "peer" + string(rune('0'+i)),
			Address:  "127.0.0.1:808" + string(rune('0'+i)),
		}
		manager.AddPeer(peer)
	}

	// Test broadcast peer selection
	peers := manager.GetOptimalPeersForBroadcast("peer0", 3)
	assert.Len(t, peers, 3)
	
	// Should not include the excluded peer
	for _, peerID := range peers {
		assert.NotEqual(t, "peer0", peerID)
	}
}