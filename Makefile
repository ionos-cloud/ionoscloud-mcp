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
	@awk 'BEGIN {FS = ":.*?## "} /^## [a-zA-Z0-9_-]+:.*?## / {sub(/^## /, ""); printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## build:    ## Build binary (VERSION=<tag> for custom version)
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o ionoscloud-mcp .

## install:  ## Install the binary to $GOBIN 
install:
	go install -trimpath -ldflags "$(LDFLAGS)" .

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

## lint:     ## Run golangci-lint (read-only)
lint:
	golangci-lint run --timeout=5m

## lintfix:  ## Run golangci-lint with --fix (auto-fixes issues)
lintfix:
	golangci-lint run --timeout=5m --fix

## vuln:     ## Run govulncheck against all packages
vuln:
	govulncheck ./...

## docker:   ## Build Docker image from source (IMAGE= to override tag)
docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

check: fmt vet lintfix vuln
	@echo "\n===>\n✓ check success!"
