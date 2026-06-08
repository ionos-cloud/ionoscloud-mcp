# syntax=docker/dockerfile:1.7

# ─── Build stage ──────────────────────────────────────────────────────────────
# Used for standalone `docker build` (developer / local images). The GoReleaser
# pipeline bypasses this stage and copies the pre-built binary directly into
# the runtime image — see .goreleaser.yml.
FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.serverVersion=${VERSION}" \
    -o /out/ionoscloud-mcp .

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="ionoscloud-mcp"
LABEL org.opencontainers.image.description="Model Context Protocol server for IONOS Cloud"
LABEL org.opencontainers.image.source="https://github.com/ionos-cloud/ionoscloud-mcp"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="IONOS Cloud"
LABEL io.modelcontextprotocol.server.name="io.github.ionos-cloud/ionoscloud-mcp"

COPY --from=build /out/ionoscloud-mcp /ionoscloud-mcp

USER nonroot:nonroot

ENTRYPOINT ["/ionoscloud-mcp"]
