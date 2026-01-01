package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

// BootstrapManager handles connections to bootstrap nodes
type BootstrapManager struct {
	nodes      []string
	connected  map[string]bool
	mu         sync.RWMutex
	maxRetries int
	retryDelay time.Duration
}

// NewBootstrapManager creates a new bootstrap manager
func NewBootstrapManager(nodes []string) *BootstrapManager {
	return &BootstrapManager{
		nodes:      nodes,
		connected:  make(map[string]bool),
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}
}

// AddNode adds a bootstrap node to the list
func (b *BootstrapManager) AddNode(node string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for _, n := range b.nodes {
		if n == node {
			return // Already exists
		}
	}
	b.nodes = append(b.nodes, node)
}

// GetNodes returns all bootstrap nodes
func (b *BootstrapManager) GetNodes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	nodes := make([]string, len(b.nodes))
	copy(nodes, b.nodes)
	return nodes
}

// ConnectToBootstrapNodes attempts to connect to all bootstrap nodes
func (b *BootstrapManager) ConnectToBootstrapNodes(ctx context.Context, connectFunc func(string) error) error {
	b.mu.RLock()
	nodes := make([]string, len(b.nodes))
	copy(nodes, b.nodes)
	b.mu.RUnlock()

	var lastErr error
	for _, node := range nodes {
		if err := b.connectWithRetry(ctx, node, connectFunc); err != nil {
			lastErr = err
			continue
		}
	}

	return lastErr
}

// connectWithRetry attempts to connect to a node with retry logic
func (b *BootstrapManager) connectWithRetry(ctx context.Context, node string, connectFunc func(string) error) error {
	var lastErr error
	
	for i := 0; i < b.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := connectFunc(node); err != nil {
			lastErr = err
			if i < b.maxRetries-1 {
				time.Sleep(b.retryDelay)
				continue
			}
		} else {
			// Mark as connected
			b.mu.Lock()
			b.connected[node] = true
			b.mu.Unlock()
			return nil
		}
	}

	return fmt.Errorf("failed to connect to bootstrap node %s after %d attempts: %w", node, b.maxRetries, lastErr)
}

// IsConnected returns whether we're connected to a specific bootstrap node
func (b *BootstrapManager) IsConnected(node string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected[node]
}

// GetConnectedNodes returns all currently connected bootstrap nodes
func (b *BootstrapManager) GetConnectedNodes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	var connectedNodes []string
	for node, isConnected := range b.connected {
		if isConnected {
			connectedNodes = append(connectedNodes, node)
		}
	}
	return connectedNodes
}

// PeerExchange handles exchanging peer information with connected nodes
type PeerExchange struct {
	maxPeers      int
	peerDiscovery func() ([]Peer, error)
	peerConnect   func(Peer) error
}

// NewPeerExchange creates a new peer exchange manager
func NewPeerExchange(maxPeers int) *PeerExchange {
	return &PeerExchange{
		maxPeers: maxPeers,
	}
}

// SetDiscoveryFunc sets the function to discover peers from connected nodes
func (p *PeerExchange) SetDiscoveryFunc(discoveryFunc func() ([]Peer, error)) {
	p.peerDiscovery = discoveryFunc
}

// SetConnectFunc sets the function to connect to discovered peers
func (p *PeerExchange) SetConnectFunc(connectFunc func(Peer) error) {
	p.peerConnect = connectFunc
}

// ExchangePeers exchanges peer information with connected nodes
func (p *PeerExchange) ExchangePeers(ctx context.Context) error {
	if p.peerDiscovery == nil || p.peerConnect == nil {
		return fmt.Errorf("discovery and connect functions must be set")
	}

	peers, err := p.peerDiscovery()
	if err != nil {
		return fmt.Errorf("failed to discover peers: %w", err)
	}

	connectedCount := 0
	for _, peer := range peers {
		if connectedCount >= p.maxPeers {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := p.peerConnect(peer); err != nil {
			// Log error but continue with other peers
			continue
		}
		connectedCount++
	}

	return nil
}

// DiscoverLocalPeers uses mDNS to discover local peers
func DiscoverLocalPeers(ctx context.Context, timeout time.Duration) ([]Peer, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	var peers []Peer
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start browsing
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := resolver.Browse(ctx, ServiceName, "local.", entries)
		if err != nil {
			return
		}
	}()

	// Process discovered entries
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-entries:
				peer := processServiceEntry(entry)
				if peer != nil {
					mu.Lock()
					peers = append(peers, *peer)
					mu.Unlock()
				}
			}
		}
	}()

	// Wait for timeout or context cancellation
	<-ctx.Done()
	wg.Wait()

	return peers, nil
}

// processServiceEntry converts a service entry to a Peer
func processServiceEntry(entry *zeroconf.ServiceEntry) *Peer {
	if len(entry.AddrIPv4) == 0 && len(entry.AddrIPv6) == 0 {
		return nil
	}

	var address string
	if len(entry.AddrIPv4) > 0 {
		address = entry.AddrIPv4[0].String()
	} else {
		address = entry.AddrIPv6[0].String()
	}

	// Extract node ID from TXT records if available
	var nodeID string
	for _, txt := range entry.Text {
		if txtParts := splitNodeID(txt); len(txtParts) == 2 && txtParts[0] == "node_id" {
			nodeID = txtParts[1]
			break
		}
	}

	return &Peer{
		ID:       nodeID,
		Address:  address,
		Port:     entry.Port,
		Hostname: entry.HostName,
		TTL:      time.Duration(entry.TTL) * time.Second,
	}
}

// splitNodeID splits a "key=value" string
func splitNodeID(s string) []string {
	parts := make([]string, 2)
	for i, c := range s {
		if c == '=' {
			parts[0] = s[:i]
			parts[1] = s[i+1:]
			return parts
		}
	}
	return []string{}
}