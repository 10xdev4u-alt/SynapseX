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
- Make (optional, for build automation)

### Installation

```bash
# Clone the repository
git clone https://github.com/princetheprogrammer/synapse.git
cd synapse

# Install dependencies
go mod download

# Build using Make (recommended)
make build

# Or build directly with Go
go build -o bin/synapse ./cmd/synapse

# Run a node
./bin/synapse
```

### Building from Source

```bash
# Format code
make fmt

# Run tests
make test

# Run tests with coverage
make coverage

# Build for all platforms
make build-all

# Development mode (debug logging)
make dev
```

### Configuration

Create a configuration file at `~/.synapse/config.json` or use command-line flags:

```bash
# Show version
./bin/synapse --version

# Specify custom config
./bin/synapse --config /path/to/config.json

# Override settings
./bin/synapse --port 9090 --log-level debug --log-format console
```

Example configuration:
```json
{
  "node": {
    "name": "my-synapse-node"
  },
  "p2p": {
    "listen_port": 8080,
    "max_peers": 50
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

See `config.example.json` for a complete configuration template.

### Running Tests

```bash
# Run all tests
make test

# Verbose output
make test-v

# Generate coverage report
make coverage
# Opens coverage.html in browser
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

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Quick Start for Contributors

1. Fork and clone the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes following [development guidelines](docs/DEVELOPMENT.md)
4. Run tests: `make test`
5. Commit with conventional commits: `feat(scope): description`
6. Push and create a Pull Request

### Development Resources

- [Architecture Documentation](docs/ARCHITECTURE.md) - System design and components
- [Development Guidelines](docs/DEVELOPMENT.md) - Coding standards and best practices
- [Phase Documentation](docs/PHASE_1.md) - Current development phase details