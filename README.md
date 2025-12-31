# Synapse

A decentralized, peer-to-peer second brain system designed for distributed knowledge management across multiple devices.

## Overview

Synapse is a distributed knowledge management system that enables seamless synchronization of notes, AI conversations, and insights across multiple devices without requiring a central server. Built on peer-to-peer architecture, it ensures your data remains under your control while being accessible from anywhere.

## Architecture

### Core Principles

- **Decentralized**: No single point of failure; each node operates independently
- **Distributed**: Data is replicated across all connected nodes
- **Version-Controlled**: Git-inspired versioning for all knowledge base changes
- **Offline-First**: Full functionality without internet connectivity
- **Privacy-Focused**: End-to-end encryption for all synchronized data

### System Design

Synapse operates as a peer-to-peer network where each device runs an independent node. These nodes:

- Generate unique identifiers for network participation
- Communicate via TCP-based P2P protocol
- Maintain local copies of the knowledge base
- Synchronize changes using conflict-free replication
- Integrate with AI services for enhanced insights

## Project Structure

```
synapse/
├── cmd/
│   └── synapse/          # Main application entry point
├── pkg/
│   ├── node/             # Core node implementation
│   ├── p2p/              # Peer-to-peer networking
│   ├── storage/          # Data persistence layer
│   ├── sync/             # Synchronization protocols
│   ├── ai/               # AI integration
│   └── crypto/           # Encryption utilities
├── internal/
│   ├── config/           # Configuration management
│   └── logger/           # Logging infrastructure
└── docs/                 # Documentation
```

## Development Phases

The project is being developed in 10 structured phases:

1. **Foundation & Research** - Architecture definition and tech stack selection
2. **Core Node Infrastructure** - Basic node with unique identity
3. **P2P Communication Layer** - Inter-node messaging
4. **Data Storage Layer** - Local persistence with content-addressing
5. **AI Integration** - External AI API connectivity
6. **Data Synchronization** - Cross-node data replication
7. **Version Control System** - Git-like versioning for data
8. **Enhanced CLI & UX** - Professional command-line interface
9. **Mobile & Multi-Device Optimization** - Resource-efficient operation
10. **Advanced Features & Polish** - Production readiness

## Technology Stack

- **Language**: Go (cross-platform, efficient concurrency)
- **P2P Networking**: Custom TCP (evolving to libp2p)
- **Storage**: BadgerDB (embedded key-value store)
- **Serialization**: JSON (transitioning to Protocol Buffers)
- **Encryption**: Go crypto/nacl
- **CLI Framework**: Cobra + Bubbletea

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/princetheprogrammer/synapse.git
cd synapse

# Build the application
go build -o bin/synapse ./cmd/synapse

# Run a node
./bin/synapse
```

## Roadmap

- [x] Phase 1: Foundation & Research
- [ ] Phase 2: Core Node Infrastructure
- [ ] Phase 3: P2P Communication Layer
- [ ] Phase 4: Data Storage Layer
- [ ] Phase 5: AI Integration
- [ ] Phase 6: Data Synchronization
- [ ] Phase 7: Version Control System
- [ ] Phase 8: Enhanced CLI & UX
- [ ] Phase 9: Mobile & Multi-Device Optimization
- [ ] Phase 10: Advanced Features & Polish

## License

MIT License - See LICENSE file for details

## Contributing

This is a personal project currently in active development. Contributions, ideas, and feedback are welcome.
