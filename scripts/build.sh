#!/bin/bash

set -e

echo "=== Synapse Build and Test Script ==="
echo ""

echo "Step 1: Format code"
gofmt -s -w .
echo "✓ Code formatted"
echo ""

echo "Step 2: Tidy dependencies"
go mod tidy
echo "✓ Dependencies tidied"
echo ""

echo "Step 3: Run tests"
go test -race -cover ./...
echo "✓ All tests passed"
echo ""

echo "Step 4: Build binary"
mkdir -p bin
VERSION=${VERSION:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
    -o bin/synapse ./cmd/synapse

echo "✓ Binary built: bin/synapse"
echo ""

echo "Step 5: Verify binary"
./bin/synapse --version
echo "✓ Binary verified"
echo ""

echo "=== Build Complete ==="
echo ""
echo "Run the application with: ./bin/synapse"
echo "Or use: make run"
