package monitor

import (
	"sync"
	"time"

	"github.com/princetheprogrammer/synapse/pkg/p2p/topology"
)

// Stats holds network statistics
type Stats struct {
	TotalMessagesSent     uint64
	TotalMessagesReceived uint64
	TotalBytesSent        uint64
	TotalBytesReceived    uint64
	ConnectionCount       int
	ActiveConnections     int
	Uptime                time.Duration
	StartTime             time.Time
	mu                    sync.RWMutex
}

// NewStats creates a new statistics instance
func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
	}
}

// IncrementMessagesSent increments the sent message counter
func (s *Stats) IncrementMessagesSent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalMessagesSent++
}

// IncrementMessagesReceived increments the received message counter
func (s *Stats) IncrementMessagesReceived() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalMessagesReceived++
}

// AddBytesSent adds to the sent bytes counter
func (s *Stats) AddBytesSent(bytes uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalBytesSent += bytes
}

// AddBytesReceived adds to the received bytes counter
func (s *Stats) AddBytesReceived(bytes uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalBytesReceived += bytes
}

// SetConnectionCount sets the total connection count
func (s *Stats) SetConnectionCount(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ConnectionCount = count
}

// SetActiveConnections sets the active connection count
func (s *Stats) SetActiveConnections(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActiveConnections = count
}

// GetStats returns a copy of the current statistics
func (s *Stats) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stats := *s
	stats.Uptime = time.Since(s.StartTime)
	return stats
}

// QualityMonitor monitors connection quality
type QualityMonitor struct {
	peers      map[string]*topology.ConnectionQuality
	mu         sync.RWMutex
	updateFunc func(string) (topology.ConnectionQuality, error)
}

// NewQualityMonitor creates a new quality monitor
func NewQualityMonitor() *QualityMonitor {
	return &QualityMonitor{
		peers: make(map[string]*topology.ConnectionQuality),
	}
}

// SetUpdateFunc sets the function to update connection quality
func (q *QualityMonitor) SetUpdateFunc(updateFunc func(string) (topology.ConnectionQuality, error)) {
	q.updateFunc = updateFunc
}

// UpdatePeerQuality updates the quality metrics for a peer
func (q *QualityMonitor) UpdatePeerQuality(peerID string, quality topology.ConnectionQuality) {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	q.peers[peerID] = &quality
}

// GetPeerQuality returns the quality metrics for a peer
func (q *QualityMonitor) GetPeerQuality(peerID string) (*topology.ConnectionQuality, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	quality, exists := q.peers[peerID]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent external modification
	qCopy := *quality
	return &qCopy, true
}

// GetAllPeerQualities returns all peer qualities
func (q *QualityMonitor) GetAllPeerQualities() map[string]topology.ConnectionQuality {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	result := make(map[string]topology.ConnectionQuality)
	for id, quality := range q.peers {
		result[id] = *quality
	}
	return result
}

// HealthChecker performs network health checks
type HealthChecker struct {
	peers       map[string]time.Time
	healthCheck func(string) bool
	mu          sync.RWMutex
	interval    time.Duration
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		peers:    make(map[string]time.Time),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// SetHealthCheckFunc sets the function to check peer health
func (h *HealthChecker) SetHealthCheckFunc(healthCheckFunc func(string) bool) {
	h.healthCheck = healthCheckFunc
}

// AddPeer adds a peer to be monitored
func (h *HealthChecker) AddPeer(peerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers[peerID] = time.Now()
}

// RemovePeer removes a peer from monitoring
func (h *HealthChecker) RemovePeer(peerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.peers, peerID)
}

// CheckPeerHealth checks the health of a specific peer
func (h *HealthChecker) CheckPeerHealth(peerID string) bool {
	if h.healthCheck == nil {
		return true // Assume healthy if no check function
	}
	return h.healthCheck(peerID)
}

// Start begins periodic health checks
func (h *HealthChecker) Start() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-h.stopCh:
				return
			case <-ticker.C:
				h.performHealthChecks()
			}
		}
	}()
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	close(h.stopCh)
	h.wg.Wait()
}

// performHealthChecks performs health checks on all peers
func (h *HealthChecker) performHealthChecks() {
	h.mu.RLock()
	peers := make([]string, 0, len(h.peers))
	for peerID := range h.peers {
		peers = append(peers, peerID)
	}
	h.mu.RUnlock()
	
	for _, peerID := range peers {
		if !h.CheckPeerHealth(peerID) {
			// Peer is unhealthy, could trigger removal or other actions
			// For now, just log
		}
	}
}

// GetUnhealthyPeers returns a list of unhealthy peers
func (h *HealthChecker) GetUnhealthyPeers() []string {
	h.mu.RLock()
	peers := make([]string, 0, len(h.peers))
	for peerID := range h.peers {
		peers = append(peers, peerID)
	}
	h.mu.RUnlock()
	
	unhealthy := []string{}
	for _, peerID := range peers {
		if !h.CheckPeerHealth(peerID) {
			unhealthy = append(unhealthy, peerID)
		}
	}
	
	return unhealthy
}

// BandwidthLimiter manages bandwidth usage
type BandwidthLimiter struct {
	maxUploadSpeed   float64 // in Mbps
	maxDownloadSpeed float64 // in Mbps
	currentUpload    float64
	currentDownload  float64
	mu               sync.RWMutex
}

// NewBandwidthLimiter creates a new bandwidth limiter
func NewBandwidthLimiter(maxUpload, maxDownload float64) *BandwidthLimiter {
	return &BandwidthLimiter{
		maxUploadSpeed:   maxUpload,
		maxDownloadSpeed: maxDownload,
	}
}

// UpdateUploadSpeed updates the current upload speed
func (b *BandwidthLimiter) UpdateUploadSpeed(speed float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.currentUpload = speed
}

// UpdateDownloadSpeed updates the current download speed
func (b *BandwidthLimiter) UpdateDownloadSpeed(speed float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.currentDownload = speed
}

// GetUploadSpeed returns the current upload speed
func (b *BandwidthLimiter) GetUploadSpeed() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentUpload
}

// GetDownloadSpeed returns the current download speed
func (b *BandwidthLimiter) GetDownloadSpeed() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentDownload
}

// IsUploadLimited returns whether upload is being limited
func (b *BandwidthLimiter) IsUploadLimited() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentUpload > b.maxUploadSpeed
}

// IsDownloadLimited returns whether download is being limited
func (b *BandwidthLimiter) IsDownloadLimited() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentDownload > b.maxDownloadSpeed
}

// GetUploadLimit returns the upload speed limit
func (b *BandwidthLimiter) GetUploadLimit() float64 {
	return b.maxUploadSpeed
}

// GetDownloadLimit returns the download speed limit
func (b *BandwidthLimiter) GetDownloadLimit() float64 {
	return b.maxDownloadSpeed
}

// NetworkMonitor combines all monitoring components
type NetworkMonitor struct {
	Stats         *Stats
	Quality       *QualityMonitor
	Health        *HealthChecker
	Bandwidth     *BandwidthLimiter
	Topology      *topology.Manager
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(topologyManager *topology.Manager) *NetworkMonitor {
	return &NetworkMonitor{
		Stats:    NewStats(),
		Quality:  NewQualityMonitor(),
		Health:   NewHealthChecker(30 * time.Second),
		Bandwidth: NewBandwidthLimiter(10.0, 10.0), // 10 Mbps default
		Topology: topologyManager,
	}
}

// Start begins all monitoring services
func (n *NetworkMonitor) Start() {
	n.Health.Start()
}

// Stop stops all monitoring services
func (n *NetworkMonitor) Stop() {
	n.Health.Stop()
}

// GetNetworkReport returns a comprehensive network report
func (n *NetworkMonitor) GetNetworkReport() map[string]interface{} {
	return map[string]interface{}{
		"stats":          n.Stats.GetStats(),
		"peer_qualities": n.Quality.GetAllPeerQualities(),
		"unhealthy_peers": n.Health.GetUnhealthyPeers(),
		"bandwidth": map[string]interface{}{
			"upload": map[string]interface{}{
				"current": n.Bandwidth.GetUploadSpeed(),
				"limit":   n.Bandwidth.GetUploadLimit(),
				"limited": n.Bandwidth.IsUploadLimited(),
			},
			"download": map[string]interface{}{
				"current": n.Bandwidth.GetDownloadSpeed(),
				"limit":   n.Bandwidth.GetDownloadLimit(),
				"limited": n.Bandwidth.IsDownloadLimited(),
			},
		},
		"topology_metrics": n.Topology.GetNetworkMetrics(),
	}
}
