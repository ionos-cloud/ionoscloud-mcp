.PHONY: build run clean test fmt vet

# Build the binary
build:
	go build -o ionoscloud-mcp .

# Run the server
run: build
	./ionoscloud-mcp

# Clean build artifacts
clean:
	rm -f ionoscloud-mcp

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
