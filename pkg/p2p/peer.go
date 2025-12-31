package p2p

import (
	"net"
	"sync"
	"time"
)

// Connection represents a connection to a peer
type Connection struct {
	ID        string
	PeerID    string
	Address   string
	Conn      net.Conn
	CreatedAt time.Time
	LastSeen  time.Time
	mu        sync.RWMutex
}

// UpdateLastSeen updates the last seen timestamp
func (c *Connection) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastSeen = time.Now()
}

// IsActive checks if the connection is still active based on timeout
func (c *Connection) IsActive(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastSeen) < timeout
}

// Peer represents a peer in the network
type Peer struct {
	ID          string
	Address     string
	Version     string
	LastSeen    time.Time
	ConnectedAt time.Time
	Connection  *Connection
	mu          sync.RWMutex
}

// NewPeer creates a new peer instance
func NewPeer(id, address, version string) *Peer {
	return &Peer{
		ID:          id,
		Address:     address,
		Version:     version,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
}

// UpdateLastSeen updates the last seen timestamp
func (p *Peer) UpdateLastSeen() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LastSeen = time.Now()
}

// IsAlive checks if the peer is still alive based on timeout
func (p *Peer) IsAlive(timeout time.Duration) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return time.Since(p.LastSeen) < timeout
}

// GetConnection returns the peer's connection
func (p *Peer) GetConnection() *Connection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Connection
}

// SetConnection sets the peer's connection
func (p *Peer) SetConnection(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Connection = conn
}
