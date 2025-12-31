.PHONY: all build test clean install run help

BINARY_NAME=synapse
BUILD_DIR=bin
CMD_DIR=cmd/synapse
PKG_DIRS=$(shell go list ./... | grep -v /vendor/)

VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

all: clean test build

help:
	@echo "Synapse Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the synapse binary"
	@echo "  test        - Run all tests"
	@echo "  test-v      - Run tests with verbose output"
	@echo "  coverage    - Generate test coverage report"
	@echo "  clean       - Remove build artifacts"
	@echo "  install     - Install binary to GOPATH/bin"
	@echo "  run         - Build and run the application"
	@echo "  fmt         - Format code with gofmt"
	@echo "  lint        - Run golangci-lint (requires installation)"
	@echo "  deps        - Download dependencies"
	@echo "  tidy        - Tidy go.mod and go.sum"
	@echo "  all         - Clean, test, and build"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Cross-platform builds complete"

test:
	@echo "Running tests..."
	go test -race -cover $(PKG_DIRS)

test-v:
	@echo "Running tests (verbose)..."
	go test -v -race -cover $(PKG_DIRS)

coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out $(PKG_DIRS)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./$(CMD_DIR)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@echo "Format complete"

lint:
	@echo "Running linter..."
	golangci-lint run ./...

deps:
	@echo "Downloading dependencies..."
	go mod download

tidy:
	@echo "Tidying go.mod and go.sum..."
	go mod tidy

dev: build
	@echo "Running in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) --log-level debug --log-format console
