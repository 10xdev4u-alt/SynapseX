package node

import (
	"context"
	"testing"
	"time"

	"github.com/princetheprogrammer/synapse/internal/config"
	"github.com/princetheprogrammer/synapse/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestNode(t *testing.T) *Node {
	cfg := config.Default()
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)

	node, err := New(cfg, log)
	require.NoError(t, err)
	return node
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		log       *logger.Logger
		expectErr bool
	}{
		{
			name:      "valid configuration",
			cfg:       config.Default(),
			log:       mustCreateLogger(t),
			expectErr: false,
		},
		{
			name:      "nil config",
			cfg:       nil,
			log:       mustCreateLogger(t),
			expectErr: true,
		},
		{
			name:      "nil logger",
			cfg:       config.Default(),
			log:       nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := New(tt.cfg, tt.log)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
				assert.NotEmpty(t, node.ID())
			}
		})
	}
}

func TestNodeID(t *testing.T) {
	node := createTestNode(t)
	
	id := node.ID()
	assert.NotEmpty(t, id)
	assert.Len(t, id, 36)
}

func TestNodeStatus(t *testing.T) {
	node := createTestNode(t)

	assert.Equal(t, StatusStopped, node.Status())
	assert.Equal(t, "stopped", node.Status().String())
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusStopped, "stopped"},
		{StatusStarting, "starting"},
		{StatusRunning, "running"},
		{StatusStopping, "stopping"},
		{Status(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestNodeStartStop(t *testing.T) {
	node := createTestNode(t)
	ctx := context.Background()

	err := node.Start(ctx)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, node.Status())

	time.Sleep(100 * time.Millisecond)

	err = node.Stop()
	require.NoError(t, err)
	assert.Equal(t, StatusStopped, node.Status())
}

func TestNodeStartTwice(t *testing.T) {
	node := createTestNode(t)
	ctx := context.Background()

	err := node.Start(ctx)
	require.NoError(t, err)

	err = node.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	node.Stop()
}

func TestNodeStopNotRunning(t *testing.T) {
	node := createTestNode(t)

	err := node.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestNodeContextCancellation(t *testing.T) {
	node := createTestNode(t)
	ctx, cancel := context.WithCancel(context.Background())

	err := node.Start(ctx)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, node.Status())

	cancel()

	done := make(chan struct{})
	go func() {
		node.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("node did not stop after context cancellation")
	}
}

func TestNodeWait(t *testing.T) {
	node := createTestNode(t)
	ctx := context.Background()

	err := node.Start(ctx)
	require.NoError(t, err)

	stopped := make(chan struct{})
	go func() {
		node.Wait()
		close(stopped)
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-stopped:
		t.Fatal("Wait() returned before Stop()")
	default:
	}

	node.Stop()

	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() did not return after Stop()")
	}
}

func TestNodeIDPersistence(t *testing.T) {
	cfg := config.Default()
	log := mustCreateLogger(t)

	node1, err := New(cfg, log)
	require.NoError(t, err)
	id1 := node1.ID()

	cfg.Node.ID = id1
	node2, err := New(cfg, log)
	require.NoError(t, err)
	id2 := node2.ID()

	assert.Equal(t, id1, id2)
}

func TestNodeInvalidID(t *testing.T) {
	cfg := config.Default()
	cfg.Node.ID = "invalid-uuid"
	log := mustCreateLogger(t)

	_, err := New(cfg, log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid node ID")
}

func mustCreateLogger(t *testing.T) *logger.Logger {
	log, err := logger.New("debug", "json", "")
	require.NoError(t, err)
	return log
}
