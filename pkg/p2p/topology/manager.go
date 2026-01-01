package topology

import (
	"math"
	"sort"
	"sync"
	"time"
)

// ConnectionQuality represents the quality of a connection
type ConnectionQuality struct {
	Latency    time.Duration
	Bandwidth  float64 // in Mbps
	PacketLoss float64 // percentage
	Jitter     time.Duration
	LastUpdate time.Time
}

// PeerInfo contains information about a peer for topology decisions
type PeerInfo struct {
	ID         string
	Address    string
	Quality    ConnectionQuality
	LastSeen   time.Time
	Connected  bool
	Reputation float64 // -1.0 to 1.0 scale
	Load       int     // number of active connections through this peer
}

// Manager handles network topology management and routing decisions
type Manager struct {
	maxPeers      int
	meshThreshold int
	peers         map[string]*PeerInfo
	mu            sync.RWMutex
	qualityUpdate func(string) ConnectionQuality
}

// NewManager creates a new topology manager
func NewManager(maxPeers int) *Manager {
	return &Manager{
		maxPeers:      maxPeers,
		meshThreshold: 10, // Switch to partial mesh after 10 peers
		peers:         make(map[string]*PeerInfo),
	}
}

// SetQualityUpdateFunc sets the function to update connection quality
func (t *Manager) SetQualityUpdateFunc(qualityFunc func(string) ConnectionQuality) {
	t.qualityUpdate = qualityFunc
}

// Peer represents a network peer
type Peer struct {
	ID       string
	Address  string
	Version  string
	LastSeen time.Time
}

// AddPeer adds a peer to the topology
func (t *Manager) AddPeer(peer Peer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	info := &PeerInfo{
		ID:         peer.ID,
		Address:    peer.Address,
		LastSeen:   time.Now(),
		Connected:  true,
		Reputation: 0.0,
		Load:       0,
	}

	// Initialize with default quality
	info.Quality = ConnectionQuality{
		Latency:    time.Second,
		Bandwidth:  1.0,
		PacketLoss: 0.0,
		Jitter:     time.Millisecond * 10,
		LastUpdate: time.Now(),
	}

	t.peers[peer.ID] = info
}

// RemovePeer removes a peer from the topology
func (t *Manager) RemovePeer(peerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.peers, peerID)
}

// UpdatePeerQuality updates the quality metrics for a peer
func (t *Manager) UpdatePeerQuality(peerID string, quality ConnectionQuality) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if peer, exists := t.peers[peerID]; exists {
		peer.Quality = quality
		peer.LastSeen = time.Now()
	}
}

// UpdatePeerReputation updates the reputation of a peer
func (t *Manager) UpdatePeerReputation(peerID string, reputation float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if peer, exists := t.peers[peerID]; exists {
		peer.Reputation = reputation
	}
}

// GetBestPeers returns the top N peers based on quality and reputation
func (t *Manager) GetBestPeers(n int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Create a slice of all peers with their scores
	type peerScore struct {
		id    string
		score float64
	}
	
	peerScores := make([]peerScore, 0, len(t.peers))
	
	for id, info := range t.peers {
		// Calculate score based on quality and reputation
		qualityScore := t.calculateQualityScore(info.Quality)
		score := qualityScore*0.7 + info.Reputation*0.3 // Weight quality more than reputation
		peerScores = append(peerScores, peerScore{id: id, score: score})
	}
	
	// Sort by score (descending)
	sort.Slice(peerScores, func(i, j int) bool {
		return peerScores[i].score > peerScores[j].score
	})
	
	// Return top n peers
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(peerScores); i++ {
		result = append(result, peerScores[i].id)
	}
	
	return result
}

// calculateQualityScore calculates a normalized quality score from connection metrics
func (t *Manager) calculateQualityScore(quality ConnectionQuality) float64 {
	// Normalize metrics to 0-1 scale
	latencyScore := 1.0 / (1.0 + float64(quality.Latency)/float64(time.Second)) // Lower latency is better
	bandwidthScore := math.Min(quality.Bandwidth/100.0, 1.0) // Cap at 1.0
	packetLossScore := 1.0 - math.Min(quality.PacketLoss/100.0, 1.0) // Lower packet loss is better
	jitterScore := 1.0 / (1.0 + float64(quality.Jitter)/float64(time.Second)) // Lower jitter is better
	
	// Weighted average
	totalScore := latencyScore*0.3 + bandwidthScore*0.3 + packetLossScore*0.2 + jitterScore*0.2
	return math.Min(totalScore, 1.0) // Cap at 1.0
}

// GetTopologyType returns the current network topology type based on peer count
func (t *Manager) GetTopologyType() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	peerCount := len(t.peers)
	
	if peerCount <= 3 {
		return "star" // Small network
	} else if peerCount <= t.meshThreshold {
		return "full-mesh" // Medium network
	} else {
		return "partial-mesh" // Large network
	}
}

// GetRoute determines the best route for a message
func (t *Manager) GetRoute(targetPeerID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	// For now, return direct route if peer exists
	if _, exists := t.peers[targetPeerID]; exists {
		return []string{targetPeerID}
	}
	
	// In the future, implement more sophisticated routing algorithms
	// like shortest path, gossip-based routing, etc.
	return nil
}

// GetPeerInfo returns information about a specific peer
func (t *Manager) GetPeerInfo(peerID string) (*PeerInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	peer, exists := t.peers[peerID]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent external modification
	info := *peer
	return &info, true
}

// GetConnectedPeers returns all connected peers
func (t *Manager) GetConnectedPeers() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	peers := make([]string, 0, len(t.peers))
	for id, info := range t.peers {
		if info.Connected {
			peers = append(peers, id)
		}
	}
	return peers
}

// GetPeerCount returns the number of known peers
func (t *Manager) GetPeerCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.peers)
}

// GetNetworkMetrics returns overall network metrics
func (t *Manager) GetNetworkMetrics() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	totalPeers := len(t.peers)
	connectedPeers := 0
	avgLatency := time.Duration(0)
	totalBandwidth := 0.0
	
	for _, info := range t.peers {
		if info.Connected {
			connectedPeers++
		}
		avgLatency += info.Quality.Latency
		totalBandwidth += info.Quality.Bandwidth
	}
	
	if totalPeers > 0 {
		avgLatency = avgLatency / time.Duration(totalPeers)
	}
	
	avgBandwidth := 0.0
	if connectedPeers > 0 {
		avgBandwidth = totalBandwidth / float64(connectedPeers)
	}
	
	return map[string]interface{}{
		"total_peers":      totalPeers,
		"connected_peers":  connectedPeers,
		"topology_type":    t.GetTopologyType(),
		"avg_latency":      avgLatency,
		"avg_bandwidth":    avgBandwidth,
		"max_peers":        t.maxPeers,
	}
}

// UpdatePeerLoad updates the load metric for a peer
func (t *Manager) UpdatePeerLoad(peerID string, load int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if peer, exists := t.peers[peerID]; exists {
		peer.Load = load
	}
}

// GetOptimalPeersForBroadcast returns the optimal set of peers for message broadcasting
func (t *Manager) GetOptimalPeersForBroadcast(excludePeerID string, maxPeers int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	// Get best peers excluding the sender
	bestPeers := t.GetBestPeers(len(t.peers))
	
	result := make([]string, 0, maxPeers)
	for _, peerID := range bestPeers {
		if peerID != excludePeerID && len(result) < maxPeers {
			if peer, exists := t.peers[peerID]; exists && peer.Connected {
				result = append(result, peerID)
			}
		}
	}
	
	return result
}