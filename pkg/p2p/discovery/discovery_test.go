package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapManager(t *testing.T) {
	nodes := []string{"192.168.1.1:8080", "192.168.1.2:8080"}
	manager := NewBootstrapManager(nodes)

	// Test initial nodes
	assert.Equal(t, nodes, manager.GetNodes())

	// Test adding a new node
	manager.AddNode("192.168.1.3:8080")
	updatedNodes := manager.GetNodes()
	assert.Len(t, updatedNodes, 3)
	assert.Contains(t, updatedNodes, "192.168.1.3:8080")
}

func TestPeerExchange(t *testing.T) {
	pe := NewPeerExchange(10)

	// Test basic creation
	assert.Equal(t, 10, pe.maxPeers)

	// Test that discovery and connect functions must be set
	err := pe.ExchangePeers(context.Background())
	assert.Error(t, err)
}

func TestDiscoverLocalPeers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// This test will likely return no peers in a test environment
	peers, err := DiscoverLocalPeers(ctx, 1*time.Second)
	assert.NoError(t, err)
	// In a test environment, we may not discover any peers
	_ = peers
}

func TestPeerExchangeExchangePeers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	pe := NewPeerExchange(5)
	
	// Set up mock discovery and connect functions
	pe.SetDiscoveryFunc(func() ([]Peer, error) {
		return []Peer{
			{ID: "peer1", Address: "127.0.0.1", Port: 8081},
			{ID: "peer2", Address: "127.0.0.1", Port: 8082},
		}, nil
	})
	
	connectCount := 0
	pe.SetConnectFunc(func(p Peer) error {
		connectCount++
		return nil
	})

	err := pe.ExchangePeers(ctx)
	assert.NoError(t, err)
	assert.Greater(t, connectCount, 0)
}