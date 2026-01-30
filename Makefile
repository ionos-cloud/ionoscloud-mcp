.PHONY: build run clean test fmt vet

VERSION ?= 0.2.0
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o ionoscloud-mcp ./cmd/ionoscloud-mcp

# Run the server
run: build
	./ionoscloud-mcp

# Clean build artifacts
clean:
	rm -f ionoscloud-mcp ionoscloud-mcp-new

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Check code quality
check: fmt vet

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build and run
dev: check build run

# List available toolsets
list-toolsets: build
	./ionoscloud-mcp list-toolsets

# List all tools
list-tools: build
	./ionoscloud-mcp list-tools
