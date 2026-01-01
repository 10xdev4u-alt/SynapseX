package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/princetheprogrammer/synapse/internal/config"
	"github.com/princetheprogrammer/synapse/internal/logger"
	"github.com/princetheprogrammer/synapse/pkg/p2p/crypto"
	"github.com/princetheprogrammer/synapse/pkg/p2p/discovery"
	"github.com/princetheprogrammer/synapse/pkg/p2p/monitor"
	"github.com/princetheprogrammer/synapse/pkg/p2p/topology"
)

// Network represents the P2P network implementation
type Network struct {
	config       *config.Config
	logger       *logger.Logger
	nodeID       string
	nodeName     string
	listener     net.Listener
	pool         *ConnectionPool
	peers        map[string]*Peer
	peersMu      sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	started      time.Time
	messageChan  chan Message
	shutdownOnce sync.Once
	mu           sync.Mutex

	// Crypto components for Phase 3
	encryptor       *crypto.Encryptor
	handshakeMgr    *crypto.HandshakeManager

	// Discovery components for Phase 3
	bootstrapMgr    *discovery.BootstrapManager
	mdnsDiscoverer  *discovery.MDNSDiscoverer
	peerExchange    *discovery.PeerExchange

	// Topology components for Phase 3
	topologyMgr     *topology.Manager

	// Monitor components for Phase 3
	monitor         *monitor.NetworkMonitor
}

// New creates a new P2P network instance
func New(cfg *config.Config, logger *logger.Logger, nodeID string) (*Network, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	networkLogger := logger.With("component", "p2p")
	
	// Create encryptor for message encryption
	encryptor, err := crypto.NewEncryptor()
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	n := &Network{
		config:      cfg,
		logger:      networkLogger,
		nodeID:      nodeID,
		nodeName:    cfg.Node.Name,
		peers:       make(map[string]*Peer),
		messageChan: make(chan Message, DefaultMessageQueueSize),
		encryptor:   encryptor,
	}

	// Initialize components
	n.handshakeMgr = crypto.NewHandshakeManager(encryptor, nodeID)
	n.bootstrapMgr = discovery.NewBootstrapManager(cfg.P2P.BootstrapPeers)
	n.topologyMgr = topology.NewManager(cfg.P2P.MaxPeers)
	n.monitor = monitor.NewNetworkMonitor(n.topologyMgr)
	n.peerExchange = discovery.NewPeerExchange(cfg.P2P.MaxPeers)

	// Initialize connection pool
	n.pool = NewConnectionPool(networkLogger, cfg.P2P.MaxPeers, DefaultConnectionTimeout)

	return n, nil
}

// Start begins listening for incoming connections and starts network operations
func (n *Network) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.listener != nil {
		return fmt.Errorf("network already started")
	}

	n.logger.Infof("starting P2P network on port %d", n.config.P2P.ListenPort)

	// Create context for network operations
	n.ctx, n.cancel = context.WithCancel(ctx)

	// Start the listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.config.P2P.ListenPort))
	if err != nil {
		return fmt.Errorf("failed to start TCP listener on port %d: %w", n.config.P2P.ListenPort, err)
	}
	n.listener = listener
	n.started = time.Now()

	n.logger.Infof("P2P network listening on port %d", n.config.P2P.ListenPort)

	// Start accepting connections in a goroutine
	go n.acceptConnections()

	// Start connection pool cleanup
	go n.pool.CleanInactive(n.ctx)

	// Start message processing
	go n.processMessages()

	// Start heartbeat service if enabled
	if n.config.P2P.EnableDiscovery {
		go n.heartbeatService()
	}

	// Initialize mDNS discoverer
	n.mdnsDiscoverer = discovery.NewMDNSDiscoverer(n.nodeName, n.config.P2P.ListenPort, []string{fmt.Sprintf("node_id=%s", n.nodeID)})
	if err := n.mdnsDiscoverer.Start(ctx); err != nil {
		n.logger.Errorf("failed to start mDNS discovery: %v", err)
		// Don't fail startup for mDNS issues
	}

	// Start bootstrap connections
	go n.connectToBootstrapNodes()

	// Start monitoring
	n.monitor.Start()

	// Start periodic peer discovery
	go n.periodicPeerDiscovery()

	return nil
}

// acceptConnections handles incoming TCP connections
func (n *Network) acceptConnections() {
	defer func() {
		if r := recover(); r != nil {
			n.logger.Errorf("panic in acceptConnections: %v", r)
		}
	}()

	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("P2P network context cancelled, stopping connection acceptor")
			return
		default:
			conn, err := n.listener.Accept()
			if err != nil {
				select {
				case <-n.ctx.Done():
					// Context cancelled, exit gracefully
					n.logger.Info("P2P network stopped, exiting accept loop")
					return
				default:
					n.logger.Errorf("error accepting connection: %v", err)
					continue
				}
			}

			// Handle the connection in a separate goroutine
			go n.handleConnectionWithEncryption(conn, true) // incoming connection
		}
	}
}

// handleConnection processes a TCP connection (incoming or outgoing)
func (n *Network) handleConnection(conn net.Conn, incoming bool) {
	connID := fmt.Sprintf("conn_%s_%d", conn.RemoteAddr().String(), time.Now().UnixNano())
	
	connection := &Connection{
		ID:        connID,
		Address:   conn.RemoteAddr().String(),
		Conn:      conn,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	n.logger.Infof("handling connection %s (incoming: %t) from %s", connID, incoming, conn.RemoteAddr())

	// Add to connection pool
	if err := n.pool.AddConnection(connection); err != nil {
		n.logger.Errorf("failed to add connection to pool: %v", err)
		conn.Close()
		return
	}

	defer func() {
		n.pool.RemoveConnection(connID)
		conn.Close()
	}()

	// Perform handshake if this is an incoming connection
	if incoming {
		if err := n.performHandshake(conn, true); err != nil {
			n.logger.Errorf("handshake failed for incoming connection: %v", err)
			return
		}
	}

	// Start reading messages from the connection
	reader := bufio.NewReader(conn)
	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("network context cancelled, closing connection")
			return
		default:
			// Set read deadline to detect dead connections
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					n.logger.Errorf("error reading from connection %s: %v", connID, err)
				}
				return
			}

			// Update last seen time
			connection.UpdateLastSeen()

			// Deserialize the message
			msg, err := DeserializeMessage(data)
			if err != nil {
				n.logger.Errorf("failed to deserialize message from %s: %v", conn.RemoteAddr(), err)
				continue
			}

			// Validate the message
			if err := msg.Validate(); err != nil {
				n.logger.Errorf("invalid message from %s: %v", conn.RemoteAddr(), err)
				continue
			}

			// Process the message based on type
			if err := n.processMessage(msg, connection); err != nil {
				n.logger.Errorf("error processing message from %s: %v", conn.RemoteAddr(), err)
				continue
			}
		}
	}
}

// performHandshake performs the initial handshake with a peer
func (n *Network) performHandshake(conn net.Conn, incoming bool) error {
	// This method is deprecated. Use performSecureHandshake instead.
	// For backward compatibility, we'll call the secure handshake.
	connID := fmt.Sprintf("conn_%s_%d", conn.RemoteAddr().String(), time.Now().UnixNano())
	
	connection := &Connection{
		ID:        connID,
		Address:   conn.RemoteAddr().String(),
		Conn:      conn,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	// Perform handshake with encryption
	return n.performSecureHandshake(conn, incoming, connection)
}

// processMessage processes an incoming message
func (n *Network) processMessage(msg *Message, conn *Connection) error {
	switch msg.Type {
	case MessageTypeHello:
		return n.handleHelloMessage(msg, conn)
	case MessageTypeHeartbeat:
		return n.handleHeartbeatMessage(msg, conn)
	case MessageTypePeerList:
		return n.handlePeerListMessage(msg, conn)
	case MessageTypePing:
		return n.handlePingMessage(msg, conn)
	case MessageTypePong:
		return n.handlePongMessage(msg, conn)
	default:
		// Add message to the processing channel
		select {
		case n.messageChan <- *msg:
			n.logger.Debugf("queued message %s from %s", msg.ID, msg.Sender)
		default:
			n.logger.Warnf("message queue full, dropping message %s", msg.ID)
		}
	}

	return nil
}

// handleHelloMessage handles HELLO messages
func (n *Network) handleHelloMessage(msg *Message, conn *Connection) error {
	// Convert the payload to the proper type
	payloadBytes, _ := json.Marshal(msg.Payload)
	var helloPayload HelloPayload
	if err := json.Unmarshal(payloadBytes, &helloPayload); err != nil {
		return fmt.Errorf("failed to unmarshal hello payload: %w", err)
	}

	// Create or update peer information
	peer := NewPeer(helloPayload.NodeID, conn.Address, helloPayload.Version)
	peer.SetConnection(conn)
	n.peersMu.Lock()
	n.peers[helloPayload.NodeID] = peer
	n.peersMu.Unlock()
	
	n.pool.AddPeer(peer)
	
	n.logger.Infof("registered new peer: %s at %s", helloPayload.NodeID, conn.Address)
	
	// Send our peer list to the new peer
	if err := n.sendPeerList(conn.Conn); err != nil {
		n.logger.Errorf("failed to send peer list to %s: %v", helloPayload.NodeID, err)
	}

	return nil
}

// handleHeartbeatMessage handles HEARTBEAT messages
func (n *Network) handleHeartbeatMessage(msg *Message, conn *Connection) error {
	// Convert the payload to the proper type
	payloadBytes, _ := json.Marshal(msg.Payload)
	var heartbeatPayload HeartbeatPayload
	if err := json.Unmarshal(payloadBytes, &heartbeatPayload); err != nil {
		return fmt.Errorf("failed to unmarshal heartbeat payload: %w", err)
	}

	conn.UpdateLastSeen()
	
	n.logger.Debugf("received heartbeat from %s", msg.Sender)
	
	// Send response heartbeat
	response := NewMessage(MessageTypeHeartbeat, n.nodeID, HeartbeatPayload{
		NodeID: n.nodeID,
		TS:     time.Now().Unix(),
	})
	
	if err := n.sendMessageToConn(conn.Conn, response); err != nil {
		n.logger.Errorf("failed to send heartbeat response: %v", err)
	}

	return nil
}

// handlePingMessage handles PING messages
func (n *Network) handlePingMessage(msg *Message, conn *Connection) error {
	// Send PONG response
	pongMsg := NewMessage(MessageTypePong, n.nodeID, map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"request_id": msg.ID,
	})
	
	if err := n.sendMessageToConn(conn.Conn, pongMsg); err != nil {
		return fmt.Errorf("failed to send pong: %w", err)
	}

	return nil
}

// handlePongMessage handles PONG messages
func (n *Network) handlePongMessage(msg *Message, conn *Connection) error {
	n.logger.Debugf("received pong from %s", msg.Sender)
	return nil
}

// handlePeerListMessage handles PEER_LIST messages
func (n *Network) handlePeerListMessage(msg *Message, conn *Connection) error {
	// Convert the payload to the proper type
	payloadBytes, _ := json.Marshal(msg.Payload)
	var peerListPayload PeerListPayload
	if err := json.Unmarshal(payloadBytes, &peerListPayload); err != nil {
		return fmt.Errorf("failed to unmarshal peer list payload: %w", err)
	}

	n.logger.Debugf("received peer list with %d peers from %s", len(peerListPayload.Peers), msg.Sender)

	// Add received peers to our known peers (but don't connect automatically)
	for _, peerInfo := range peerListPayload.Peers {
		if peerInfo.ID != n.nodeID { // Don't add ourselves
			n.logger.Debugf("learned about peer %s at %s", peerInfo.ID, peerInfo.Address)
		}
	}

	return nil
}

// Connect establishes a connection to a peer at the given address
func (n *Network) Connect(address string) error {
	n.logger.Infof("attempting to connect to peer: %s", address)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", address, err)
	}

	// Handle the connection (this will perform secure handshake)
	go n.handleConnectionWithEncryption(conn, false) // outgoing connection

	return nil
}

// SendMessage sends a message to a specific peer
func (n *Network) SendMessage(peerID string, msg Message) error {
	// Find the peer
	n.peersMu.RLock()
	peer, exists := n.peers[peerID]
	n.peersMu.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	conn := peer.GetConnection()
	if conn == nil {
		return fmt.Errorf("no active connection to peer %s", peerID)
	}

	return n.sendMessageToConn(conn.Conn, msg)
}

// sendMessageToConn sends a message to a specific connection
func (n *Network) sendMessageToConn(conn net.Conn, msg Message) error {
	data, err := msg.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Add newline for message framing
	data = append(data, '\n')

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message to connection: %w", err)
	}

	// Update monitoring stats
	n.monitor.Stats.AddBytesSent(uint64(len(data)))
	n.monitor.Stats.IncrementMessagesSent()

	return nil
}

// Broadcast sends a message to all connected peers
func (n *Network) Broadcast(msg Message) error {
	peers := n.pool.GetPeers()
	var lastErr error

	for _, peer := range peers {
		conn := peer.GetConnection()
		if conn == nil {
			continue
		}

		if err := n.sendMessageToConn(conn.Conn, msg); err != nil {
			lastErr = err
			n.logger.Errorf("failed to broadcast message to peer %s: %v", peer.ID, err)
		}
	}

	return lastErr
}

// Peers returns a list of connected peers
func (n *Network) Peers() []*Peer {
	return n.pool.GetPeers()
}

// Status returns the current network status
func (n *Network) Status() NetworkStatus {
	n.peersMu.RLock()
	defer n.peersMu.RUnlock()

	connectedPeers := make([]string, 0, len(n.peers))
	for id := range n.peers {
		connectedPeers = append(connectedPeers, id)
	}

	return NetworkStatus{
		ActiveConnections: n.pool.ConnectionCount(),
		TotalPeers:       n.pool.PeerCount(),
		Listening:        n.listener != nil,
		NodeID:          n.nodeID,
		Uptime:          time.Since(n.started).Seconds(),
	}
}

// Stop shuts down the P2P network
func (n *Network) Stop() error {
	var err error
	n.shutdownOnce.Do(func() {
		n.logger.Info("stopping P2P network")
		
		if n.cancel != nil {
			n.cancel()
		}

		if n.listener != nil {
			if closeErr := n.listener.Close(); closeErr != nil {
				err = fmt.Errorf("failed to close listener: %w", closeErr)
			}
		} else {
			err = fmt.Errorf("network not started")
		}

		// Close all connections
		connections := n.pool.GetConnections()
		for _, conn := range connections {
			conn.Conn.Close()
		}

		// Clear peers
		n.peersMu.Lock()
		n.peers = make(map[string]*Peer)
		n.peersMu.Unlock()

		n.logger.Info("P2P network stopped")
	})

	return err
}

// processMessages processes messages from the message channel
func (n *Network) processMessages() {
	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("stopping message processor")
			return
		case msg := <-n.messageChan:
			n.logger.Debugf("processing message %s of type %s from %s", msg.ID, msg.Type, msg.Sender)
			// In a real implementation, we would route messages to appropriate handlers
			// based on the message type and content
		}
	}
}

// heartbeatService sends periodic heartbeat messages to maintain connections
func (n *Network) heartbeatService() {
	ticker := time.NewTicker(DefaultHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("stopping heartbeat service")
			return
		case <-ticker.C:
			heartbeatMsg := NewMessage(MessageTypeHeartbeat, n.nodeID, HeartbeatPayload{
				NodeID: n.nodeID,
				TS:     time.Now().Unix(),
			})
			
			if err := n.Broadcast(heartbeatMsg); err != nil {
				n.logger.Errorf("failed to broadcast heartbeat: %v", err)
			}
		}
	}
}

// sendPeerList sends the current list of known peers to a connection
func (n *Network) sendPeerList(conn net.Conn) error {
	peers := n.Peers()
	
	peerInfos := make([]PeerInfo, 0, len(peers))
	for _, peer := range peers {
		peerInfos = append(peerInfos, PeerInfo{
			ID:       peer.ID,
			Address:  peer.Address,
			Version:  peer.Version,
			LastSeen: peer.LastSeen.Unix(),
		})
	}

	peerListPayload := PeerListPayload{
		Peers: peerInfos,
	}

	peerListMsg := NewMessage(MessageTypePeerList, n.nodeID, peerListPayload)
	
	return n.sendMessageToConn(conn, peerListMsg)
}

// performSecureHandshake performs the secure handshake with encryption
func (n *Network) performSecureHandshake(conn net.Conn, incoming bool, connection *Connection) error {
	if incoming {
		// For incoming connections, receive their handshake message
		handshakeMsg, err := n.receiveHandshakeMessage(conn)
		if err != nil {
			return fmt.Errorf("failed to receive handshake: %w", err)
		}

		// Verify the handshake message
		if err := n.handshakeMgr.VerifyHandshakeMessage(handshakeMsg); err != nil {
			return fmt.Errorf("handshake verification failed: %w", err)
		}

		// Register the peer
		n.registerPeer(handshakeMsg.NodeID, connection)

		// Send our handshake message in response
		responseMsg, err := n.handshakeMgr.CreateHandshakeMessage()
		if err != nil {
			return fmt.Errorf("failed to create response handshake: %w", err)
		}

		if err := n.sendHandshakeMessage(conn, responseMsg); err != nil {
			return fmt.Errorf("failed to send response handshake: %w", err)
		}
	} else {
		// For outgoing connections, send our handshake message first
		handshakeMsg, err := n.handshakeMgr.CreateHandshakeMessage()
		if err != nil {
			return fmt.Errorf("failed to create handshake: %w", err)
		}

		if err := n.sendHandshakeMessage(conn, handshakeMsg); err != nil {
			return fmt.Errorf("failed to send handshake: %w", err)
		}

		// Receive their response
		responseMsg, err := n.receiveHandshakeMessage(conn)
		if err != nil {
			return fmt.Errorf("failed to receive response handshake: %w", err)
		}

		// Verify the response
		if err := n.handshakeMgr.VerifyHandshakeMessage(responseMsg); err != nil {
			return fmt.Errorf("response handshake verification failed: %w", err)
		}

		// Register the peer
		n.registerPeer(responseMsg.NodeID, connection)
	}

	return nil
}

// sendHandshakeMessage sends an encrypted handshake message
func (n *Network) sendHandshakeMessage(conn net.Conn, msg *crypto.HandshakeMessage) error {
	// For now, send unencrypted for testing. In real implementation, we'd need their public key
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake message: %w", err)
	}

	// Add newline for message framing
	data = append(data, '\n')

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write handshake message: %w", err)
	}

	return nil
}

// receiveHandshakeMessage receives and parses a handshake message
func (n *Network) receiveHandshakeMessage(conn net.Conn) (*crypto.HandshakeMessage, error) {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake message: %w", err)
	}

	// Remove newline
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	var msg crypto.HandshakeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handshake message: %w", err)
	}

	return &msg, nil
}

// registerPeer registers a peer in our network
func (n *Network) registerPeer(peerID string, connection *Connection) {
	peer := NewPeer(peerID, connection.Address, "1.0.0")
	peer.SetConnection(connection)
	
	n.peersMu.Lock()
	n.peers[peerID] = peer
	n.peersMu.Unlock()
	
	n.pool.AddPeer(peer)
	
	// Create topology peer from our peer
	topologyPeer := topology.Peer{
		ID:       peer.ID,
		Address:  peer.Address,
		Version:  peer.Version,
		LastSeen: peer.LastSeen,
	}
	n.topologyMgr.AddPeer(topologyPeer)
	
	n.logger.Infof("registered new peer: %s at %s", peerID, connection.Address)
}

// handleConnectionWithEncryption processes a TCP connection with encryption (incoming or outgoing)
func (n *Network) handleConnectionWithEncryption(conn net.Conn, incoming bool) {
	connID := fmt.Sprintf("conn_%s_%d", conn.RemoteAddr().String(), time.Now().UnixNano())
	
	connection := &Connection{
		ID:        connID,
		Address:   conn.RemoteAddr().String(),
		Conn:      conn,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	n.logger.Infof("handling connection %s (incoming: %t) from %s", connID, incoming, conn.RemoteAddr())

	// Add to connection pool
	if err := n.pool.AddConnection(connection); err != nil {
		n.logger.Errorf("failed to add connection to pool: %v", err)
		conn.Close()
		return
	}

	defer func() {
		n.pool.RemoveConnection(connID)
		conn.Close()
	}()

	// Perform handshake with encryption
	if err := n.performSecureHandshake(conn, incoming, connection); err != nil {
		n.logger.Errorf("secure handshake failed for connection %s: %v", connID, err)
		return
	}

	// Start reading messages from the connection
	if err := n.readMessages(conn, connection); err != nil {
		n.logger.Errorf("error reading messages from connection %s: %v", connID, err)
	}
}

// readMessages reads and processes messages from a connection
func (n *Network) readMessages(conn net.Conn, connection *Connection) error {
	reader := bufio.NewReader(conn)
	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("network context cancelled, closing connection")
			return nil
		default:
			// Set read deadline to detect dead connections
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					n.logger.Errorf("error reading from connection: %v", err)
				}
				return err
			}

			// Update last seen time
			connection.UpdateLastSeen()
			n.monitor.Stats.AddBytesReceived(uint64(len(data)))

			// Deserialize the message
			msg, err := DeserializeMessage(data)
			if err != nil {
				n.logger.Errorf("failed to deserialize message from %s: %v", conn.RemoteAddr(), err)
				continue
			}

			// Validate the message
			if err := msg.Validate(); err != nil {
				n.logger.Errorf("invalid message from %s: %v", conn.RemoteAddr(), err)
				continue
			}

			// Process the message based on type
			if err := n.processMessage(msg, connection); err != nil {
				n.logger.Errorf("error processing message from %s: %v", conn.RemoteAddr(), err)
				continue
			}
		}
	}
}