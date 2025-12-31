package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/princetheprogrammer/synapse/internal/config"
	"github.com/princetheprogrammer/synapse/internal/logger"
)

type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusStopping
)

func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

type Node struct {
	id     string
	config *config.Config
	logger *logger.Logger
	status Status
	mu     sync.RWMutex

	stopCh chan struct{}
	doneCh chan struct{}
}

func New(cfg *config.Config, log *logger.Logger) (*Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	nodeID := cfg.Node.ID
	if nodeID == "" {
		nodeID = uuid.New().String()
		cfg.Node.ID = nodeID
	}

	if _, err := uuid.Parse(nodeID); err != nil {
		return nil, fmt.Errorf("invalid node ID format: %w", err)
	}

	return &Node{
		id:     nodeID,
		config: cfg,
		logger: log.With("node_id", nodeID),
		status: StatusStopped,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}, nil
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Status() Status {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}

func (n *Node) setStatus(status Status) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.status = status
	n.logger.Infof("node status changed to: %s", status)
}

func (n *Node) Start(ctx context.Context) error {
	if n.Status() != StatusStopped {
		return fmt.Errorf("node already running or starting")
	}

	n.setStatus(StatusStarting)
	n.logger.Info("starting synapse node")

	if err := n.initialize(); err != nil {
		n.setStatus(StatusStopped)
		return fmt.Errorf("failed to initialize node: %w", err)
	}

	go n.run(ctx)

	n.setStatus(StatusRunning)
	n.logger.Infof("synapse node started successfully on port %d", n.config.P2P.ListenPort)

	return nil
}

func (n *Node) initialize() error {
	n.logger.Debug("initializing node components")
	return nil
}

func (n *Node) run(ctx context.Context) {
	defer close(n.doneCh)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			n.logger.Info("context cancelled, shutting down")
			return

		case <-n.stopCh:
			n.logger.Info("stop signal received, shutting down")
			return

		case <-ticker.C:
			n.logger.Debug("node heartbeat")
		}
	}
}

func (n *Node) Stop() error {
	if n.Status() != StatusRunning {
		return fmt.Errorf("node is not running")
	}

	n.setStatus(StatusStopping)
	n.logger.Info("stopping synapse node")

	close(n.stopCh)

	shutdownTimeout := time.NewTimer(10 * time.Second)
	defer shutdownTimeout.Stop()

	select {
	case <-n.doneCh:
		n.logger.Info("node stopped gracefully")
	case <-shutdownTimeout.C:
		n.logger.Warn("node shutdown timeout, forcing stop")
	}

	n.setStatus(StatusStopped)
	return nil
}

func (n *Node) Wait() {
	<-n.doneCh
}
