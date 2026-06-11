.PHONY: help build install run clean test cover test-e2e fmt vet check deps dev lint lintfix vuln docker snapshot

.DEFAULT_GOAL := help

VERSION ?= dev
LDFLAGS := -s -w -X main.serverVersion=$(VERSION)
IMAGE   ?= ghcr.io/ionos-cloud/ionoscloud-mcp:$(VERSION)

## help:     ## Show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^## [a-zA-Z_-]+:.*?## / {sub(/^## /, ""); printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## build:    ## Build the binary (VERSION=<tag> to override version string)
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o ionoscloud-mcp .

## install:  ## Install the binary to $GOBIN (or $GOPATH/bin) so MCP clients on PATH pick it up
install:
	go install -trimpath -ldflags "$(LDFLAGS)" .

## run:      ## Build and run the MCP server
run: build
	./ionoscloud-mcp

## clean:    ## Remove build artifacts and dist/
clean:
	rm -f ionoscloud-mcp
	rm -rf dist/

## test:     ## Run unit tests
test:
	go test -v -race ./...

## cover:    ## Run tests with cross-package coverage (writes coverage.out)
cover:
	go test -race -coverpkg=./... -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

## test-e2e: ## Binary-over-stdio (mocked API) + read-only live API checks (live skips without IONOS_TOKEN)
test-e2e:
	go test -tags e2e_live -count=1 -timeout 20m ./test/live/...
	go test -race -tags e2e -count=1 ./test/e2e/...

## fmt:      ## Format Go code with gofmt
fmt:
	go fmt ./...

## vet:      ## Run go vet
vet:
	go vet ./...

## check:    ## Run fmt + vet
check: fmt vet

## lint:     ## Run golangci-lint (read-only)
lint:
	golangci-lint run --timeout=5m

## lintfix:  ## Run golangci-lint with --fix (auto-fixes issues)
lintfix:
	golangci-lint run --timeout=5m --fix

## vuln:     ## Run govulncheck against all packages
vuln:
	govulncheck ./...

## docker:   ## Build a local Docker image from source (IMAGE= to override tag)
docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

## snapshot: ## Dry-run the GoReleaser pipeline locally (no publish)
snapshot:
	goreleaser release --snapshot --clean --skip=publish

## deps:     ## go mod download + tidy
deps:
	go mod download
	go mod tidy

## dev:      ## check + build + run
dev: check build run
