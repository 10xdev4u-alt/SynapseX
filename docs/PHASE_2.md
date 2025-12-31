# Phase 2: Core Node Infrastructure

**Status**: In Progress  
**Goal**: Build the core P2P communication layer for Synapse nodes

## Objectives

- [x] Research TCP-based P2P protocols and connection management
- [x] Design message protocol for node-to-node communication  
- [x] Plan peer discovery and connection pooling strategies
- [x] Define message types and serialization format
- [ ] Implement basic TCP server functionality
- [ ] Implement TCP client for peer connections
- [ ] Create message serialization/deserialization
- [ ] Build connection pool management
- [ ] Add heartbeat/keep-alive mechanism
- [ ] Implement peer discovery logic
- [ ] Add connection error handling and retry logic
- [ ] Create message routing system
- [ ] Build peer status monitoring
- [ ] Add security considerations for P2P communication
- [ ] Implement basic message validation
- [ ] Create network statistics and monitoring
- [ ] Add configurable network timeouts
- [ ] Implement graceful connection handling
- [ ] Add peer connection limits
- [ ] Create network configuration options

## Research Findings

### TCP-based P2P Networking

**Decision**: Custom TCP protocol with future migration to libp2p

**Rationale**:
- Learning fundamentals before advanced libraries
- Full control over protocol design
- Simpler for initial implementation
- libp2p integration planned for Phase 6+

**Protocol Design**:
- JSON-based message format for development speed
- Length-prefixed messages to handle framing
- Message types for different operations
- Error handling for network failures

### Message Protocol

**Format**:
```json
{
  "type": "message_type",
  "id": "unique_message_id",
  "sender": "sender_node_id",
  "timestamp": "RFC3339_timestamp",
  "payload": {}
}
```

**Message Types**:
- `HELLO`: Initial connection handshake
- `PEER_LIST`: Exchange known peers
- `DATA_SYNC`: Synchronize data changes
- `HEARTBEAT`: Keep-alive messages
- `ERROR`: Error notifications

### Connection Management

**Approaches Evaluated**:

1. **Connection Pooling**
   - Pro: Efficient resource utilization
   - Pro: Reuse existing connections
   - Pro: Better performance under load
   - Implementation: Maintain active connections to peers

2. **Peer Discovery**
   - Pro: Automatic peer detection
   - Pro: Dynamic network membership
   - Implementation: Bootstrap peers + gossip protocol

3. **Heartbeat System**
   - Pro: Detect dead connections
   - Pro: Maintain network health
   - Implementation: Periodic ping/pong mechanism

## Implementation Strategy

### Package Structure

```
pkg/p2p/
├── p2p.go                 # Main P2PNetwork interface
├── network.go             # Network implementation
├── connection.go          # Connection management
├── message.go             # Message types and serialization
├── peer.go                # Peer representation
├── pool.go                # Connection pool
└── protocol.go            # Protocol constants and helpers
```

### Core Components

1. **P2PNetwork Interface**
   - `Start()` / `Stop()` for lifecycle
   - `Connect(address)` for peer connections
   - `Broadcast(msg)` for network-wide messages
   - `SendMessage(peerID, msg)` for targeted messages
   - `Peers()` for peer discovery
   - `Status()` for network health

2. **Connection Management**
   - TCP connection handling
   - Message framing and serialization
   - Error handling and recovery
   - Connection pooling

3. **Message System**
   - Type-safe message handling
   - Serialization/deserialization
   - Message validation
   - Routing logic

## Timeline

- Started: Dec 31, 2024
- Research: Dec 31, 2024
- Implementation: Dec 31, 2024
- Testing: Dec 31, 2024
- **Target Completion**: Dec 31, 2024

## Success Criteria

- [ ] TCP server accepts incoming connections
- [ ] TCP client connects to peers successfully
- [ ] Messages can be sent/received between nodes
- [ ] Connection pool manages multiple peers
- [ ] Heartbeat system detects dead connections
- [ ] Peer discovery works for new nodes
- [ ] Error handling for network failures
- [ ] Configurable network settings
- [ ] Unit tests >80% coverage
- [ ] Integration tests pass

## Dependencies

- Standard Go `net` package for TCP
- Standard Go `json` for serialization
- Existing `config` package for settings
- Existing `logger` package for monitoring

## Security Considerations

- Message validation to prevent injection
- Connection limits to prevent resource exhaustion
- Timeout handling for stuck connections
- Basic authentication in future phases

## Performance Targets

- Handle 50+ concurrent connections
- Message latency < 100ms on local network
- Connection establishment < 1s
- Minimal memory usage per connection

## Next Steps (Phase 3)

- Advanced peer discovery (mDNS, DHT)
- Message encryption
- Network topology optimization
- Advanced routing algorithms
