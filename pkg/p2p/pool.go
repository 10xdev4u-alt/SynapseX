package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	DefaultMaxConnections = 50
)

// ConnectionPool manages a pool of connections to peers
type ConnectionPool struct {
	maxConnections int
	timeout        time.Duration
	connections    map[string]*Connection
	peers          map[string]*Peer
	mu             sync.RWMutex
	logger         Logger
}

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(logger Logger, maxConnections int, timeout time.Duration) *ConnectionPool {
	if maxConnections <= 0 {
		maxConnections = DefaultMaxConnections
	}
	if timeout <= 0 {
		timeout = DefaultConnectionTimeout
	}

	return &ConnectionPool{
		maxConnections: maxConnections,
		timeout:        timeout,
		connections:    make(map[string]*Connection),
		peers:          make(map[string]*Peer),
		logger:         logger,
	}
}

// AddConnection adds a connection to the pool
func (cp *ConnectionPool) AddConnection(conn *Connection) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if len(cp.connections) >= cp.maxConnections {
		return fmt.Errorf("connection pool at maximum capacity (%d)", cp.maxConnections)
	}

	cp.connections[conn.ID] = conn
	cp.logger.Debugf("added connection %s to pool", conn.ID)
	return nil
}

// RemoveConnection removes a connection from the pool
func (cp *ConnectionPool) RemoveConnection(connID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if conn, exists := cp.connections[connID]; exists {
		conn.Conn.Close()
		delete(cp.connections, connID)
		cp.logger.Debugf("removed connection %s from pool", connID)
	}
}

// GetConnection retrieves a connection by ID
func (cp *ConnectionPool) GetConnection(connID string) (*Connection, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	conn, exists := cp.connections[connID]
	return conn, exists
}

// AddPeer adds a peer to the pool
func (cp *ConnectionPool) AddPeer(peer *Peer) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.peers[peer.ID] = peer
	cp.logger.Debugf("added peer %s to pool", peer.ID)
}

// RemovePeer removes a peer from the pool
func (cp *ConnectionPool) RemovePeer(peerID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	delete(cp.peers, peerID)
	cp.logger.Debugf("removed peer %s from pool", peerID)
}

// GetPeer retrieves a peer by ID
func (cp *ConnectionPool) GetPeer(peerID string) (*Peer, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	peer, exists := cp.peers[peerID]
	return peer, exists
}

// GetPeers returns all peers in the pool
func (cp *ConnectionPool) GetPeers() []*Peer {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	peers := make([]*Peer, 0, len(cp.peers))
	for _, peer := range cp.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetConnections returns all connections in the pool
func (cp *ConnectionPool) GetConnections() []*Connection {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	conns := make([]*Connection, 0, len(cp.connections))
	for _, conn := range cp.connections {
		conns = append(conns, conn)
	}
	return conns
}

// CleanInactive removes inactive connections from the pool
func (cp *ConnectionPool) CleanInactive(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cp.logger.Info("stopping connection pool cleanup")
			return
		case <-ticker.C:
			cp.cleanInactiveConnections()
		}
	}
}

// cleanInactiveConnections removes connections that have been inactive
func (cp *ConnectionPool) cleanInactiveConnections() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	inactive := []string{}
	for id, conn := range cp.connections {
		if !conn.IsActive(cp.timeout) {
			inactive = append(inactive, id)
		}
	}

	for _, id := range inactive {
		conn := cp.connections[id]
		conn.Conn.Close()
		delete(cp.connections, id)
		cp.logger.Infof("removed inactive connection %s", id)
	}

	if len(inactive) > 0 {
		cp.logger.Debugf("cleaned %d inactive connections", len(inactive))
	}
}

// PeerCount returns the number of peers in the pool
func (cp *ConnectionPool) PeerCount() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.peers)
}

// ConnectionCount returns the number of connections in the pool
func (cp *ConnectionPool) ConnectionCount() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.connections)
}

// IsFull checks if the connection pool is at maximum capacity
func (cp *ConnectionPool) IsFull() bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.connections) >= cp.maxConnections
}