.PHONY: build run clean test fmt vet check deps dev lint vuln docker snapshot

VERSION ?= dev
LDFLAGS := -s -w -X main.serverVersion=$(VERSION)
IMAGE   ?= ghcr.io/ionos-cloud/ionoscloud-mcp:$(VERSION)

# Build the binary
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o ionoscloud-mcp .

# Run the server
run: build
	./ionoscloud-mcp

# Clean build artifacts
clean:
	rm -f ionoscloud-mcp
	rm -rf dist/

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

# Run golangci-lint (requires golangci-lint installed)
lint:
	golangci-lint run --timeout=5m

# Run govulncheck (requires govulncheck installed: go install golang.org/x/vuln/cmd/govulncheck@latest)
vuln:
	govulncheck ./...

# Build a local Docker image from source (uses ./Dockerfile, not the release one)
docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

# Dry-run the release pipeline locally (no publish). Requires goreleaser installed.
snapshot:
	goreleaser release --snapshot --clean --skip=publish

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build and run
dev: check build run
