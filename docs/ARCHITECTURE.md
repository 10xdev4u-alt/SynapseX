# Synapse Architecture

## System Overview

Synapse is a decentralized peer-to-peer knowledge management system designed to operate across multiple devices without a central authority.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        User Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  CLI Client  │  │ Future: TUI  │  │ Future: API  │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
┌─────────┴──────────────────┴──────────────────┴─────────────┐
│                    Application Layer                        │
│  ┌────────────────────────────────────────────────────┐     │
│  │              Synapse Node Core                     │     │
│  │  - Node Identity Management                        │     │
│  │  - Command Processing                              │     │
│  │  - State Management                                │     │
│  └────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
          │
┌─────────┴───────────────────────────────────────────────────┐
│                    Service Layer                            │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐              │
│  │  P2P Net   │ │  Storage   │ │    Sync    │              │
│  │            │ │            │ │            │              │
│  │ - Discovery│ │ - BadgerDB │ │ - Merkle   │              │
│  │ - Messaging│ │ - CRUD Ops │ │ - CRDTs    │              │
│  │ - Peers    │ │ - Index    │ │ - Conflict │              │
│  └────────────┘ └────────────┘ └────────────┘              │
│                                                              │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐              │
│  │     AI     │ │   Crypto   │ │  Versioning│              │
│  │            │ │            │ │            │              │
│  │ - API Conn │ │ - Encrypt  │ │ - Events   │              │
│  │ - History  │ │ - Signing  │ │ - Time Trvl│              │
│  │ - Streaming│ │ - Verify   │ │ - History  │              │
│  └────────────┘ └────────────┘ └────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Node Core (`pkg/node`)

**Responsibilities:**
- Node lifecycle management (startup, shutdown)
- Unique node identity (UUID-based)
- Configuration loading and validation
- Health monitoring

**Key Interfaces:**
```go
type Node interface {
    Start() error
    Stop() error
    ID() string
    Status() NodeStatus
}
```

### 2. P2P Networking (`pkg/p2p`)

**Responsibilities:**
- TCP-based peer communication
- Peer discovery and management
- Message routing
- Connection pool management
- Heartbeat/keep-alive

**Evolution Path:**
- Phase 3: Custom TCP protocol
- Phase 6+: Migrate to libp2p for advanced features

**Key Interfaces:**
```go
type P2PNetwork interface {
    Listen(port string) error
    Connect(address string) error
    Broadcast(msg Message) error
    Send(peerID string, msg Message) error
    Peers() []Peer
}
```

### 3. Storage Layer (`pkg/storage`)

**Responsibilities:**
- Local data persistence
- Content-addressable storage (hash-based keys)
- Indexing and querying
- Data integrity verification

**Technology:** BadgerDB
- Embedded key-value store
- LSM-tree based (optimized for writes)
- ACID transactions
- Low memory footprint

**Data Model:**
```go
type StorageEntry struct {
    Hash      string    // Content-addressable key
    Type      EntryType // Note, AIChat, Metadata
    Content   []byte    // Encrypted payload
    Timestamp time.Time
    Version   int64
}
```

### 4. Synchronization (`pkg/sync`)

**Responsibilities:**
- Data replication across nodes
- Conflict detection and resolution
- Delta sync for efficiency
- Version vector maintenance

**Approach:**
- **Merkle Trees**: Efficient diff computation
- **CRDTs**: Conflict-free replicated data types for specific data structures
- **Event Sourcing**: Immutable log of all changes
- **Timestamp-based Conflict Resolution**: Last-write-wins (Phase 6)
- **Advanced Conflict Resolution**: User-driven or AI-assisted (Phase 10)

**Sync Protocol:**
```
1. Node A requests sync from Node B
2. Exchange Merkle tree roots
3. Identify differing branches
4. Request missing data blocks
5. Apply changes with conflict resolution
6. Update version vectors
```

### 5. AI Integration (`pkg/ai`)

**Responsibilities:**
- HTTP client for external AI API
- Request/response handling
- Conversation history management
- Streaming response support
- Rate limiting and retry logic

**API Integration:**
- Endpoint: `https://svceai.site/api/chat`
- Method: POST
- Payload: `{message: string, history: []Message}`
- Response: AI-generated text

**Features:**
- Offline queue for failed requests
- Local caching of responses
- Integration with sync layer for cross-device history

### 6. Cryptography (`pkg/crypto`)

**Responsibilities:**
- End-to-end encryption
- Digital signatures
- Key management
- Data integrity verification

**Implementation:**
- **Encryption**: NaCl (libsodium) secretbox
- **Signing**: Ed25519
- **Key Derivation**: Argon2
- **Transport Security**: TLS for P2P (Phase 6+)

### 7. Version Control (`pkg/version` - Future)

**Responsibilities:**
- Event sourcing for all changes
- Time-travel queries
- Branching/merging concepts for ideas
- Garbage collection

**Inspired by Git:**
- Immutable commit graph
- Content-addressable storage
- Diff/patch operations
- Rebase/merge strategies

## Data Flow

### Write Path (Adding a Note)

```
User Input (CLI)
    │
    ├─> Node Core (validate)
    │       │
    │       ├─> Storage Layer (persist)
    │       │       │
    │       │       └─> BadgerDB (commit)
    │       │
    │       └─> Sync Layer (propagate)
    │               │
    │               ├─> Create event log entry
    │               ├─> Update Merkle tree
    │               └─> P2P Broadcast
    │                       │
    │                       └─> Connected Peers (replicate)
    │
    └─> Return success to user
```

### Sync Path (Receiving Updates)

```
P2P Network (receive message)
    │
    ├─> Sync Layer (process)
    │       │
    │       ├─> Verify signature
    │       ├─> Check Merkle tree
    │       ├─> Detect conflicts
    │       │       │
    │       │       ├─> No conflict: Accept
    │       │       └─> Conflict: Resolve (timestamp/CRDT/user)
    │       │
    │       └─> Storage Layer (apply)
    │               │
    │               └─> BadgerDB (commit)
    │
    └─> Update local Merkle tree
```

### AI Query Path

```
User Query (CLI)
    │
    ├─> AI Layer (prepare request)
    │       │
    │       ├─> Load conversation history (Storage)
    │       ├─> Build request payload
    │       └─> HTTP POST to API
    │               │
    │               ├─> Success: Parse response
    │               └─> Failure: Queue for retry
    │
    ├─> Storage Layer (save AI response)
    │
    ├─> Sync Layer (propagate to peers)
    │
    └─> Display to user
```

## Scalability Considerations

### Network Topology

- **Full Mesh (Phase 3-6)**: All nodes connect to all peers
  - Simple implementation
  - Works for small networks (< 10 nodes)
  - O(n²) connections

- **Hybrid Topology (Phase 10+)**: Strategic peer selection
  - DHT-based peer discovery (Kademlia)
  - Connection limits per node
  - Gossip protocol for propagation

### Storage Optimization

- **Deduplication**: Content-addressable storage prevents duplicates
- **Compression**: LZ4/Snappy for large payloads
- **Pruning**: Configurable history retention
- **Archival**: Cold storage for old data

### Bandwidth Optimization

- **Delta Sync**: Only transfer changed data
- **Compression**: All network payloads compressed
- **Batching**: Aggregate small changes
- **Priority Queuing**: User-facing changes first

## Security Model

### Threat Model

**Protected Against:**
- Unauthorized data access (encryption at rest)
- Man-in-the-middle attacks (TLS + signatures)
- Data tampering (cryptographic hashing)
- Replay attacks (nonces + timestamps)

**Out of Scope (Phase 1):**
- Sophisticated network-level attacks
- Compromised node detection
- Byzantine fault tolerance

### Security Layers

1. **Transport Security**: TLS 1.3 for P2P connections
2. **Data Encryption**: NaCl secretbox for stored data
3. **Authentication**: Digital signatures for all messages
4. **Integrity**: SHA-256 hashing for content addressing

## Deployment Architecture

### Single Device

```
┌─────────────────┐
│  Synapse Node   │
│                 │
│  - Local DB     │
│  - AI Client    │
│  - No P2P       │
└─────────────────┘
```

### Multi-Device (Target)

```
┌──────────┐         ┌──────────┐
│ Laptop   │◄───────►│ Hostel PC│
│  Node A  │         │  Node B  │
└────┬─────┘         └────┬─────┘
     │                    │
     │    ┌──────────┐    │
     └───►│ College  │◄───┘
          │ PC (Node)│
          └────┬─────┘
               │
          ┌────┴─────┐
          │  Mobile  │
          │  Node D  │
          └──────────┘
```

## Future Enhancements

### Phase 7+
- IPFS integration for large file storage
- WebRTC for browser-based nodes
- Blockchain-based conflict resolution timestamps
- Federated learning for local AI models
- Plugin system for extensibility

## References

- libp2p: https://github.com/libp2p/go-libp2p
- BadgerDB: https://github.com/dgraph-io/badger
- CRDTs: https://crdt.tech/
- Event Sourcing: https://martinfowler.com/eaaDev/EventSourcing.html
