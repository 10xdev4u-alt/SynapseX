package topology

import (
	"sync"
	"time"
)

// ReputationSystem manages peer reputation based on various factors
type ReputationSystem struct {
	manager *Manager
	mu      sync.RWMutex
}

// NewReputationSystem creates a new reputation system
func NewReputationSystem(manager *Manager) *ReputationSystem {
	return &ReputationSystem{
		manager: manager,
	}
}

// UpdateReputationBasedOnBehavior updates peer reputation based on observed behavior
func (rs *ReputationSystem) UpdateReputationBasedOnBehavior(peerID string, behaviorScore float64) {
	// Validate behavior score (-1.0 to 1.0)
	if behaviorScore < -1.0 {
		behaviorScore = -1.0
	} else if behaviorScore > 1.0 {
		behaviorScore = 1.0
	}

	// Get current reputation and update it gradually
	currentInfo, exists := rs.manager.GetPeerInfo(peerID)
	if !exists {
		return
	}

	// Weighted update: 70% current reputation, 30% new behavior
	newReputation := currentInfo.Reputation*0.7 + behaviorScore*0.3
	
	// Keep reputation within bounds
	if newReputation < -1.0 {
		newReputation = -1.0
	} else if newReputation > 1.0 {
		newReputation = 1.0
	}

	rs.manager.UpdatePeerReputation(peerID, newReputation)
}

// UpdateReputationBasedOnPerformance updates peer reputation based on performance metrics
func (rs *ReputationSystem) UpdateReputationBasedOnPerformance(peerID string, successRate float64, responseTime time.Duration) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Calculate performance score based on success rate and response time
	performanceScore := successRate // successRate should be 0.0 to 1.0
	
	// Adjust based on response time (faster responses get higher scores)
	// Assume 100ms is a good response time
	maxAcceptableTime := time.Second
	if responseTime < time.Millisecond*100 {
		performanceScore *= 1.2 // Boost for fast responses
	} else if responseTime > maxAcceptableTime {
		performanceScore *= 0.8 // Penalty for slow responses
	}

	// Convert to -1.0 to 1.0 scale
	scaledScore := (performanceScore * 2.0) - 1.0
	if scaledScore < -1.0 {
		scaledScore = -1.0
	} else if scaledScore > 1.0 {
		scaledScore = 1.0
	}

	rs.UpdateReputationBasedOnBehavior(peerID, scaledScore)
}

// UpdateReputationBasedOnReliability updates peer reputation based on reliability metrics
func (rs *ReputationSystem) UpdateReputationBasedOnReliability(peerID string, uptimeRatio float64, messageDeliveryRate float64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Calculate reliability score as weighted average
	reliabilityScore := uptimeRatio*0.6 + messageDeliveryRate*0.4
	
	// Convert to -1.0 to 1.0 scale
	scaledScore := (reliabilityScore * 2.0) - 1.0
	if scaledScore < -1.0 {
		scaledScore = -1.0
	} else if scaledScore > 1.0 {
		scaledScore = 1.0
	}

	rs.UpdateReputationBasedOnBehavior(peerID, scaledScore)
}

// GetTrustedPeers returns peers with reputation above a threshold
func (rs *ReputationSystem) GetTrustedPeers(threshold float64) []string {
	// For now, return the best peers from the topology manager
	return rs.manager.GetBestPeers(10) // Return top 10 peers
}

// DecayReputation gradually reduces reputation of inactive peers
func (rs *ReputationSystem) DecayReputation(peerID string, decayRate float64) {
	currentInfo, exists := rs.manager.GetPeerInfo(peerID)
	if !exists {
		return
	}

	// Apply decay to move reputation toward neutral (0.0)
	newReputation := currentInfo.Reputation * (1 - decayRate)
	
	// Ensure it stays within bounds
	if newReputation < -1.0 {
		newReputation = -1.0
	} else if newReputation > 1.0 {
		newReputation = 1.0
	}

	rs.manager.UpdatePeerReputation(peerID, newReputation)
}

// GetPeerRank returns a rank (1-10) for a peer based on reputation
func (rs *ReputationSystem) GetPeerRank(peerID string) int {
	info, exists := rs.manager.GetPeerInfo(peerID)
	if !exists {
		return 1 // Lowest rank for unknown peers
	}

	// Convert reputation (-1.0 to 1.0) to rank (1-10)
	rank := int((info.Reputation + 1.0) * 5) // Maps -1.0 to 0, 1.0 to 10
	if rank < 1 {
		rank = 1
	} else if rank > 10 {
		rank = 10
	}

	return rank
}