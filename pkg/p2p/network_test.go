package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/princetheprogrammer/synapse/internal/config"
	"github.com/princetheprogrammer/synapse/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestNetwork(t *testing.T) (*Network, context.Context, context.CancelFunc) {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	
	network, err := New(cfg, log, "test-node-id")
	require.NoError(t, err)

	return network, ctx, cancel
}

func TestNew(t *testing.T) {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	tests := []struct {
		name      string
		cfg       *config.Config
		log       *logger.Logger
		nodeID    string
		expectErr bool
	}{
		{
			name:      "valid configuration",
			cfg:       cfg,
			log:       log,
			nodeID:    "test-node",
			expectErr: false,
		},
		{
			name:      "nil config",
			cfg:       nil,
			log:       log,
			nodeID:    "test-node",
			expectErr: true,
		},
		{
			name:      "nil logger",
			cfg:       cfg,
			log:       nil,
			nodeID:    "test-node",
			expectErr: true,
		},
		{
			name:      "empty node ID",
			cfg:       cfg,
			log:       log,
			nodeID:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network, err := New(tt.cfg, tt.log, tt.nodeID)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, network)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, network)
				assert.Equal(t, tt.nodeID, network.nodeID)
			}
		})
	}
}

func TestNetworkStartStop(t *testing.T) {
	network, ctx, cancel := createTestNetwork(t)
	defer cancel()

	err := network.Start(ctx)
	require.NoError(t, err)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	status := network.Status()
	assert.True(t, status.Listening)
	assert.Equal(t, "test-node-id", status.NodeID)
	assert.Greater(t, status.Uptime, float64(0))

	err = network.Stop()
	assert.NoError(t, err)
}

func TestNetworkStartTwice(t *testing.T) {
	network, ctx, cancel := createTestNetwork(t)
	defer cancel()

	err := network.Start(ctx)
	require.NoError(t, err)

	// Try to start again - this should fail since network is already started
	err = network.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network already started")

	err = network.Stop()
	assert.NoError(t, err)
}

func TestNetworkStopWithoutStart(t *testing.T) {
	network, _, _ := createTestNetwork(t)

	err := network.Stop()
	// In our implementation, Stop() will return an error if not started
	assert.Error(t, err)
}

func TestNetworkStatus(t *testing.T) {
	network, ctx, cancel := createTestNetwork(t)
	defer cancel()

	// Initially not listening
	status := network.Status()
	assert.False(t, status.Listening)

	err := network.Start(ctx)
	require.NoError(t, err)

	// After start, should be listening
	time.Sleep(100 * time.Millisecond)
	status = network.Status()
	assert.True(t, status.Listening)
	assert.Equal(t, "test-node-id", status.NodeID)

	err = network.Stop()
	assert.NoError(t, err)
}

func TestMessageSerialization(t *testing.T) {
	msg := NewMessage("TEST", "sender-id", map[string]interface{}{"key": "value"})

	data, err := msg.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	deserialized, err := DeserializeMessage(data)
	assert.NoError(t, err)
	assert.Equal(t, "TEST", deserialized.Type)
	assert.Equal(t, "sender-id", deserialized.Sender)
	assert.Equal(t, msg.ID, deserialized.ID)
}

func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name        string
		message     Message
		expectValid bool
	}{
		{
			name: "valid message",
			message: Message{
				Type:   "TEST",
				ID:     "test-id",
				Sender: "sender-id",
			},
			expectValid: true,
		},
		{
			name: "empty type",
			message: Message{
				Type:   "",
				ID:     "test-id",
				Sender: "sender-id",
			},
			expectValid: false,
		},
		{
			name: "empty ID",
			message: Message{
				Type:   "TEST",
				ID:     "",
				Sender: "sender-id",
			},
			expectValid: false,
		},
		{
			name: "empty sender",
			message: Message{
				Type:   "TEST",
				ID:     "test-id",
				Sender: "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage("TEST_TYPE", "test-sender", map[string]interface{}{"data": "value"})

	assert.Equal(t, "TEST_TYPE", msg.Type)
	assert.Equal(t, "test-sender", msg.Sender)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, map[string]interface{}{"data": "value"}, msg.Payload)
	assert.WithinDuration(t, time.Now(), msg.Timestamp, 1*time.Second)
}

func TestConnectionPool(t *testing.T) {
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	pool := NewConnectionPool(log, 10, 30*time.Second)

	assert.Equal(t, 0, pool.PeerCount())
	assert.Equal(t, 0, pool.ConnectionCount())
	assert.False(t, pool.IsFull())

	// Test adding a peer
	peer := NewPeer("peer-id", "127.0.0.1:8080", "1.0.0")
	pool.AddPeer(peer)
	assert.Equal(t, 1, pool.PeerCount())

	// Test getting peer
	gotPeer, exists := pool.GetPeer("peer-id")
	assert.True(t, exists)
	assert.Equal(t, "peer-id", gotPeer.ID)

	// Test getting all peers
	peers := pool.GetPeers()
	assert.Equal(t, 1, len(peers))
	assert.Equal(t, "peer-id", peers[0].ID)

	// Test removing peer
	pool.RemovePeer("peer-id")
	assert.Equal(t, 0, pool.PeerCount())
}

func TestPeer(t *testing.T) {
	peer := NewPeer("peer-id", "127.0.0.1:8080", "1.0.0")

	assert.Equal(t, "peer-id", peer.ID)
	assert.Equal(t, "127.0.0.1:8080", peer.Address)
	assert.Equal(t, "1.0.0", peer.Version)

	// Update last seen to now
	peer.UpdateLastSeen()
	assert.True(t, peer.IsAlive(10*time.Second))
	
	// Set last seen to a long time ago to test IsAlive
	peer.mu.Lock()
	peer.LastSeen = time.Now().Add(-2 * time.Minute)
	peer.mu.Unlock()
	assert.False(t, peer.IsAlive(30*time.Second))
}

func TestConnection(t *testing.T) {
	conn := &Connection{
		ID:        "test-conn",
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	assert.True(t, conn.IsActive(10*time.Second))

	// Update last seen to a long time ago
	conn.mu.Lock()
	conn.LastSeen = time.Now().Add(-2 * time.Minute)
	conn.mu.Unlock()

	assert.False(t, conn.IsActive(30*time.Second))
}