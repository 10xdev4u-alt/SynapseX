package p2p

// Interface defines the core P2P networking interface
type Interface interface {
	Start() error
	Stop() error
	Connect(address string) error
	SendMessage(peerID string, message Message) error
	Broadcast(message Message) error
	Peers() []Peer
	Status() Status
}

// Status represents the status of the P2P network
type Status struct {
	ActiveConnections int
	TotalPeers      int
	Listening       bool
	NodeID          string
	Uptime          int64
}

// NetworkStatus represents the status of the P2P network
type NetworkStatus struct {
	ActiveConnections int
	TotalPeers      int
	Listening       bool
	NodeID          string
	Uptime          float64
}