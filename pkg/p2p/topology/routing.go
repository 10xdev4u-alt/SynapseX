package topology

import (
	"math"
	"sync"
	"time"
)

// RoutingStrategy defines different routing strategies
type RoutingStrategy int

const (
	// Direct routing sends messages directly to the target
	Direct RoutingStrategy = iota
	// Gossip routing spreads messages through random peers
	Gossip
	// ShortestPath routing finds the shortest path to the target
	ShortestPath
)

// Router handles message routing decisions
type Router struct {
	manager    *Manager
	strategy   RoutingStrategy
	mu         sync.RWMutex
	routeCache map[string][]string // Cache of computed routes
}

// NewRouter creates a new router with the specified strategy
func NewRouter(manager *Manager, strategy RoutingStrategy) *Router {
	return &Router{
		manager:    manager,
		strategy:   strategy,
		routeCache: make(map[string][]string),
	}
}

// RouteMessage determines the route for a message to the target
func (r *Router) RouteMessage(targetID string) []string {
	r.mu.RLock()
	if route, exists := r.routeCache[targetID]; exists {
		r.mu.RUnlock()
		return route
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check cache again after acquiring write lock
	if route, exists := r.routeCache[targetID]; exists {
		return route
	}

	var route []string
	switch r.strategy {
	case Direct:
		route = r.directRoute(targetID)
	case Gossip:
		route = r.gossipRoute(targetID)
	case ShortestPath:
		route = r.shortestPathRoute(targetID)
	default:
		route = r.directRoute(targetID)
	}

	r.routeCache[targetID] = route
	return route
}

// directRoute returns the direct route to the target
func (r *Router) directRoute(targetID string) []string {
	// Check if the target peer exists
	_, exists := r.manager.GetPeerInfo(targetID)
	if exists {
		return []string{targetID}
	}
	return nil
}

// gossipRoute returns a route that spreads the message to random peers
func (r *Router) gossipRoute(targetID string) []string {
	// For gossip, we return a selection of well-connected peers
	bestPeers := r.manager.GetBestPeers(3)
	return bestPeers
}

// shortestPathRoute computes the shortest path to the target
// This is a simplified implementation - in a real system, this would be more complex
func (r *Router) shortestPathRoute(targetID string) []string {
	// In a real P2P network, this would use distributed routing algorithms
	// like Chord, Kademlia, etc. For now, we'll return direct route if possible
	// or route through best peers
	_, exists := r.manager.GetPeerInfo(targetID)
	if exists {
		return []string{targetID}
	}

	// If we don't know about the target, route through best peers
	bestPeers := r.manager.GetBestPeers(2)
	return bestPeers
}

// UpdateRouteCache invalidates the route cache
func (r *Router) UpdateRouteCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routeCache = make(map[string][]string)
}

// UpdatePeerMetrics updates metrics for routing decisions
func (r *Router) UpdatePeerMetrics(peerID string, latency float64, bandwidth float64) {
	quality := ConnectionQuality{
		Latency:    fromFloat64(latency),
		Bandwidth:  bandwidth,
		PacketLoss: math.Min(latency*10, 100), // Higher latency may indicate higher packet loss
	}
	r.manager.UpdatePeerQuality(peerID, quality)
}

// fromFloat64 converts a float64 to time.Duration (for testing purposes)
func fromFloat64(f float64) ConnectionQuality {
	return ConnectionQuality{
		Latency: time.Duration(f * float64(time.Millisecond)),
	}
}