package p2p

import "time"

// Protocol constants for the P2P network
const (
	// ProtocolVersion represents the current version of the P2P protocol
	ProtocolVersion = "1.0.0"
	
	// MaxMessageSize is the maximum size of a single message in bytes (1MB)
	MaxMessageSize = 1024 * 1024
	
	// MaxPeerListSize is the maximum number of peers to include in a single peer list message
	MaxPeerListSize = 100
	
	// DefaultListenPort is the default port for P2P communication
	DefaultListenPort = 8080
	
	// DefaultMaxPeers is the default maximum number of connected peers
	DefaultMaxPeers = 50
	
	// DefaultConnectionTimeout is the default timeout for connections
	DefaultConnectionTimeout = 30 * time.Second
	
	// DefaultHeartbeatInterval is the interval for sending heartbeat messages
	DefaultHeartbeatInterval = 10 * time.Second
	
	// DefaultPeerDiscoveryInterval is the interval for discovering new peers
	DefaultPeerDiscoveryInterval = 30 * time.Second
	
	// DefaultMessageQueueSize is the size of the message queue for each connection
	DefaultMessageQueueSize = 100
	
	// DefaultMaxRetries is the maximum number of retries for failed operations
	DefaultMaxRetries = 3
	
	// DefaultRetryDelay is the delay between retries
	DefaultRetryDelay = 1 * time.Second
)

// Additional message types (beyond those defined elsewhere)
const (
	// MessageTypePing is used for network latency measurement
	MessageTypePing = "PING"
	
	// MessageTypePong is used as response to ping
	MessageTypePong = "PONG"
	
	// MessageTypeSyncRequest is used to request specific data
	MessageTypeSyncRequest = "SYNC_REQUEST"
	
	// MessageTypeSyncResponse is used to respond to sync requests
	MessageTypeSyncResponse = "SYNC_RESPONSE"
)

// Capability flags for peer capabilities
const (
	// CapabilitySync indicates the peer supports data synchronization
	CapabilitySync = "sync"
	
	// CapabilityDiscovery indicates the peer supports peer discovery
	CapabilityDiscovery = "discovery"
	
	// CapabilityEncryption indicates the peer supports encrypted communication
	CapabilityEncryption = "encryption"
	
	// CapabilityRelay indicates the peer supports message relaying
	CapabilityRelay = "relay"
)

// Error codes for P2P protocol
const (
	// ErrorCodeInvalidMessage indicates an invalid message format
	ErrorCodeInvalidMessage = "INVALID_MESSAGE"
	
	// ErrorCodeConnectionFailed indicates a connection failure
	ErrorCodeConnectionFailed = "CONNECTION_FAILED"
	
	// ErrorCodePeerNotFound indicates a peer could not be found
	ErrorCodePeerNotFound = "PEER_NOT_FOUND"
	
	// ErrorCodeMaxPeersReached indicates the maximum number of peers is reached
	ErrorCodeMaxPeersReached = "MAX_PEERS_REACHED"
	
	// ErrorCodeTimeout indicates an operation timed out
	ErrorCodeTimeout = "TIMEOUT"
	
	// ErrorCodeNotImplemented indicates a feature is not implemented
	ErrorCodeNotImplemented = "NOT_IMPLEMENTED"
)