# Development Guidelines

## Code Style

### General Principles

1. **Clarity over cleverness**: Write code that is easy to understand
2. **Consistency**: Follow existing patterns in the codebase
3. **Simplicity**: Prefer simple solutions over complex ones
4. **Documentation**: Document all exported functions and types

### Go-Specific Standards

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for automatic formatting
- Run `golangci-lint` before committing
- Keep functions small and focused (ideally < 50 lines)
- Avoid global variables except for constants
- Use meaningful variable names (avoid `x`, `y`, `tmp`)

### Package Organization

```
package/
├── package.go          # Main implementation
├── package_test.go     # Unit tests
├── types.go            # Type definitions (if many)
├── errors.go           # Error definitions (if many)
└── doc.go              # Package documentation
```

### Naming Conventions

- **Packages**: Short, lowercase, single word (e.g., `node`, `storage`)
- **Interfaces**: Verb or noun ending in `-er` (e.g., `Reader`, `Syncer`)
- **Structs**: PascalCase (e.g., `NodeConfig`, `P2PNetwork`)
- **Methods/Functions**: PascalCase for exported, camelCase for private
- **Constants**: PascalCase or UPPER_CASE for groups
- **Variables**: camelCase, descriptive names

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to connect to peer %s: %w", peerID, err)
}

// Bad: Generic error messages
if err != nil {
    return err
}

// Good: Check specific error types when needed
if errors.Is(err, ErrPeerNotFound) {
    // Handle specific case
}
```

### Context Usage

- Always pass `context.Context` as first parameter for cancelable operations
- Respect context cancellation in loops and long-running operations
- Use `context.WithTimeout` for operations with time limits

```go
func (n *Node) Start(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-n.ready:
        return nil
    }
}
```

## Testing

### Test Coverage

- Maintain minimum 80% code coverage
- Focus on critical paths and edge cases
- Use table-driven tests for multiple scenarios

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    // Arrange
    setup := createTestSetup()
    
    // Act
    result, err := setup.DoSomething()
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        expected  bool
        expectErr bool
    }{
        {name: "valid input", input: "abc", expected: true, expectErr: false},
        {name: "invalid input", input: "", expected: false, expectErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Validate(tt.input)
            if tt.expectErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Mocking

- Use interfaces for dependencies to enable mocking
- Prefer `testify/mock` for complex mocking scenarios
- Keep mocks simple and focused

## Git Workflow

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `build`: Build system changes
- `perf`: Performance improvements

**Examples**:
```
feat(p2p): add peer discovery mechanism

Implement mDNS-based peer discovery for local network nodes.
This allows automatic peer detection without manual configuration.

Closes #42

---

fix(storage): resolve race condition in concurrent writes

Add mutex protection for shared state in storage layer.
Prevents data corruption during simultaneous write operations.

---

docs(arch): update synchronization strategy

Document the decision to use CRDTs for conflict resolution
instead of timestamp-based LWW approach.
```

### Branching Strategy

- `main`: Production-ready code
- `develop`: Integration branch (not used initially)
- Feature branches: Short-lived, merged quickly

### Commit Frequency

- Commit early and often (target 30+ commits per phase)
- Each commit should be a logical unit of work
- Commits should build successfully
- Keep commits focused (single responsibility)

### Before Committing

1. Run tests: `go test ./...`
2. Run linter: `golangci-lint run`
3. Format code: `go fmt ./...`
4. Review changes: `git diff`
5. Write descriptive commit message

## Code Review

### Self-Review Checklist

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] No commented-out code
- [ ] No debug prints or temporary hacks
- [ ] Error handling is comprehensive
- [ ] No unnecessary dependencies added

### Review Principles

- Be respectful and constructive
- Ask questions rather than make demands
- Explain the "why" behind suggestions
- Acknowledge good practices

## Performance

### Optimization Guidelines

1. **Profile before optimizing**: Use `pprof` to find bottlenecks
2. **Avoid premature optimization**: Correct first, fast second
3. **Benchmark critical paths**: Use `testing.B` for benchmarks
4. **Consider memory allocations**: Reuse buffers, use sync.Pool

### Benchmarking

```go
func BenchmarkOperation(b *testing.B) {
    setup := createSetup()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        setup.Operation()
    }
}
```

## Concurrency

### Guidelines

- Use goroutines sparingly and with clear purpose
- Always provide a way to stop goroutines (context, done channel)
- Protect shared state with mutexes or channels
- Prefer channels for communication, mutexes for state

### Goroutine Management

```go
// Good: Cancelable goroutine
func (n *Node) Run(ctx context.Context) error {
    errCh := make(chan error, 1)
    
    go func() {
        errCh <- n.doWork(ctx)
    }()
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    case err := <-errCh:
        return err
    }
}
```

## Documentation

### Package Documentation

```go
// Package p2p implements peer-to-peer networking for Synapse nodes.
//
// The p2p package provides functionality for:
//   - Peer discovery (mDNS, bootstrap nodes)
//   - Connection management
//   - Message routing
//   - Network health monitoring
//
// Example usage:
//   network := p2p.New(config)
//   if err := network.Start(ctx); err != nil {
//       log.Fatal(err)
//   }
package p2p
```

### Function Documentation

```go
// Connect establishes a connection to a peer at the given address.
//
// The address should be in the format "host:port". If the connection
// fails, an error wrapping the underlying cause is returned.
//
// Example:
//   err := network.Connect("192.168.1.100:8080")
func (n *Network) Connect(address string) error {
    // implementation
}
```

### Complex Logic

```go
// Calculate the Merkle root using a bottom-up approach.
// 1. Hash all leaf nodes
// 2. Combine adjacent pairs and hash
// 3. Repeat until single root hash remains
func calculateMerkleRoot(data [][]byte) []byte {
    // implementation with inline comments for tricky parts
}
```

## Dependencies

### Adding Dependencies

1. Evaluate necessity (can we implement it ourselves simply?)
2. Check license compatibility (prefer MIT, Apache 2.0, BSD)
3. Verify maintenance status (recent commits, active issues)
4. Consider size and transitive dependencies
5. Document why the dependency was added

### Preferred Libraries

- **Testing**: `testify` for assertions and mocking
- **Logging**: `zerolog` for structured logging
- **CLI**: `cobra` for command-line interface
- **TUI**: `bubbletea` for terminal UI
- **Crypto**: Standard library `crypto/*` and `golang.org/x/crypto`

## Security

### Best Practices

1. Never log sensitive data (keys, passwords, tokens)
2. Use constant-time comparisons for secrets
3. Validate all external input
4. Sanitize data before storage
5. Use parameterized queries (when SQL is added)
6. Keep dependencies updated

### Sensitive Data

```go
// Bad: Logging sensitive data
log.Infof("API key: %s", apiKey)

// Good: Redacted logging
log.Info("API key configured")

// Good: Secure comparison
if subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 {
    // Authenticated
}
```

## Project-Specific Conventions

### Node Identity

- Node IDs are UUIDs (RFC 4122)
- Always validate node ID format
- Store node ID persistently after first generation

### P2P Messages

- All messages must have a type field
- Include sender node ID
- Add timestamp for debugging
- Validate message structure before processing

### Storage Keys

- Use content-addressable keys (SHA-256 hashes)
- Prefix keys by type (e.g., `note:`, `chat:`, `meta:`)
- Never expose raw storage layer to application code

### Error Messages

- Include context (what operation, which resource)
- Use error wrapping with `%w`
- Define package-level sentinel errors for common cases

```go
var (
    ErrPeerNotFound = errors.New("peer not found")
    ErrInvalidMessage = errors.New("invalid message format")
)
```

## Continuous Improvement

- Refactor when you see duplication (DRY principle)
- Update documentation as you go
- Add TODOs for future improvements
- Learn from code reviews
- Share knowledge with team (future)

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
