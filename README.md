# IONOS CLOUD MCP Server

![Official IONOS Cloud](https://img.shields.io/badge/IONOS%20Cloud-Official-00BFFF.svg)
[![Apache 2.0](https://img.shields.io/github/license/ionos-cloud/ionoscloud-mcp)](LICENSE)
[![Go reference](https://pkg.go.dev/badge/github.com/ionos-cloud/ionoscloud-mcp.svg)](https://pkg.go.dev/github.com/ionos-cloud/ionoscloud-mcp)

A **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server that connects your IONOS CLOUD account to any MCP-compatible AI assistant or autonomous AI agent: Claude Desktop, Cursor, VS Code (GitHub Copilot), Windsurf, Cline, Continue, OpenCode, and 5+ others. **112 tools across 6 IONOS CLOUD products** — list, inspect, and audit your infrastructure through natural-language prompts or programmatic agentic loops.

Built and maintained by the IONOS Cloud team. The server runs as a local binary on your workstation, a CI runner, or inside a container. IONOS CLOUD API calls go directly to IONOS over HTTPS; no third-party AI provider sits in the data path.

**Compatibility:** MCP spec 2024-11-05 · Go 1.25+ for builds · OCI images for linux/amd64 and linux/arm64.

> 📚 **Full product documentation, per-client setup guides, FAQ, and tutorials:** [docs.ionos.com/cloud/ai/mcp-server](https://docs.ionos.com/cloud/ai/mcp-server)

**Get started in 60 seconds** (macOS or Linux, via Homebrew):

```bash
brew install ionos-cloud/ionos-cloud/ionoscloud-mcp
```

For other install paths (Docker, pre-built binary, `go install`, source), see [Installation](#installation).

<p align="center">
<a href="#why">Why</a> •
<a href="#registries--directories">Registries</a> •
<a href="#supported-products">Products</a> •
<a href="#installation">Install</a> •
<a href="#configuration">Config</a> •
<a href="#tool-loading-mode">Tool loading</a> •
<a href="#demo">Demo</a> •
<a href="#development">Dev</a> •
<a href="#related-projects">Related</a> •
<a href="#changelog">Changelog</a>
</p>

## Why

* **Read-only by design** — every tool is an inspection operation (`list_*`, `get_*`, `head_*`). The server cannot create, modify, or delete any resource, so it's safe to connect to production accounts and safe to deploy inside unattended agent loops on CI runners.
* **Local binary, no proxy** — IONOS CLOUD API calls go directly from your machine to IONOS Cloud. No third-party AI vendor in the data path.
* **EU-sovereign option** — pair the server with the [IONOS CLOUD AI Model Hub](https://docs.ionos.com/cloud/ai/ai-model-hub) and both the API calls *and* the LLM inference terminate inside IONOS's German data centres. See the [Fully Sovereign AI Workflow](https://docs.ionos.com/cloud/ai/mcp-server/use-cases/sovereign-ai-workflow) guide.
* **Open source** — Apache 2.0. Read the source, audit the behaviour, contribute, or fork.

## Registries & Directories

This server is published across multiple MCP registries and IDE marketplaces:

| Registry | Link |
|----------|------|
| Official MCP Registry | [io.github.ionos-cloud/ionoscloud-mcp](https://registry.modelcontextprotocol.io/v0/servers?cursor=io.github.ionos-cloud) |
| Smithery | [ionos-cloud/ionoscloud-mcp](https://smithery.ai/servers/ionos-cloud/ionoscloud-mcp) |
| mcp.so | [ionos-cloud-mcp-server](https://mcp.so/server/ionos-cloud-mcp-server/ionos-cloud) |
| Cursor | [ionoscloud-mcp](https://cursor.directory/plugins/ionoscloud-mcp) |
| mcpservers.org | [ionoscloud-mcp](https://mcpservers.org/servers/ionos-cloud/ionoscloud-mcp) |

## Supported products

All tools follow the `list_*`, `get_*`, and `head_*` naming convention. In `lazy` mode, two loader tools (`ionos_load_compute_tools`, `ionos_load_objectstorage_tools`) register the Compute and Object Storage catalogues on demand; in the default `eager` mode all tools register at startup. See [Tool loading mode](#tool-loading-mode).

| Product | Tools | Capabilities |
|---|---|---|
| [Compute Engine](docs/compute/) | 50 | Data centers, servers, volumes, NICs, LANs, firewall rules, IP blocks, load balancers (basic / network / application), NAT gateways, security groups, private cross-connects, snapshots, images, templates, locations, requests, contract |
| [Kubernetes](docs/k8s/) | 8 | Clusters, node pools, nodes, available versions |
| [Object Storage](docs/objectstorage/) | 23 | Buckets, bucket configuration (CORS, encryption, lifecycle, policy, public access block, replication, tagging, versioning, Object Lock), objects, access keys, regions |
| [DNS](docs/dns/) | 14 | Zones, zone files, records, reverse records, secondary zones, DNSSEC, quota |
| [Billing](docs/billing/) | 15 | Profile, invoices, EVN (provisioning intervals), traffic, usage, utilization, product pricing catalog, FOCUS v1.3 spec |
| [Certificate Manager](docs/cert/) | 6 | Certificates, auto-certificates, providers |
| [Activity Log](docs/activitylog/) | 2 | Contracts, events |

**120 tools total** (118 product + 2 loader). For per-tool input/output schemas, see the [per-product docs](docs/) or the full [Tool Reference](https://docs.ionos.com/cloud/ai/mcp-server/tool-reference) at docs.ionos.com.

## Installation

Pick whichever fits your workflow.

### Homebrew (macOS, Linux) — recommended

```bash
brew install ionos-cloud/ionos-cloud/ionoscloud-mcp
```

### Docker (linux/amd64, linux/arm64)

```bash
docker pull ghcr.io/ionos-cloud/ionoscloud-mcp:latest
```

Run with the MCP stdio transport:

```bash
docker run -i --rm \
  -e IONOS_TOKEN="$IONOS_TOKEN" \
  ghcr.io/ionos-cloud/ionoscloud-mcp
```

### Smithery

```bash
npx -y @smithery/cli install @ionos-cloud/ionoscloud-mcp --client claude-desktop
```

Supported `--client` values: `claude-desktop`, `claude-code`, `cursor`, `vscode`, `windsurf`, `cline`, `continue`, `gemini-cli`, `kiro`, and others. See the [Smithery listing](https://smithery.ai/servers/ionos-cloud/ionoscloud-mcp) for the current list.

### Pre-built binary

Download the archive for your OS/arch from the [latest release](https://github.com/ionos-cloud/ionoscloud-mcp/releases/latest). Available for Linux, macOS, and Windows on both amd64 and arm64.

### `go install`

```bash
go install github.com/ionos-cloud/ionoscloud-mcp@latest
```

### From source

```bash
git clone https://github.com/ionos-cloud/ionoscloud-mcp.git
cd ionoscloud-mcp
make build
```

## Configuration

You need an IONOS CLOUD account with API credentials.

```bash
# Required: API token for control-plane APIs (Compute, DNS, Billing, Certificate Manager, Object Storage management)
export IONOS_TOKEN="your-api-token"

# Optional: only required if you use Object Storage data-plane tools
# (listing objects, reading bucket configuration, checking access keys).
export IONOS_S3_ACCESS_KEY="your-s3-access-key"
export IONOS_S3_SECRET_KEY="your-s3-secret-key"
```

Generate a token in the [IONOS CLOUD DCD](https://dcd.ionos.com/) under **Management → Token Management**. Object Storage credentials are created under **Storage & Backup → IONOS CLOUD Object Storage → Key management**.

For least-privilege token scoping, see [Authentication](https://docs.ionos.com/cloud/ai/mcp-server/configuration/authentication) at docs.ionos.com.

### Integrating with an MCP client (manual)

Add the server to your AI client's MCP config:

```json
{
  "mcpServers": {
    "ionoscloud": {
      "command": "/path/to/ionoscloud-mcp",
      "env": {
        "IONOS_TOKEN": "your-api-token",
        "IONOS_S3_ACCESS_KEY": "your-s3-access-key",
        "IONOS_S3_SECRET_KEY": "your-s3-secret-key"
      }
    }
  }
}
```

The Object Storage credentials are only needed if you plan to use Object Storage tools.

Per-client setup guides for the 12 supported AI clients: [Connect to an AI Client](https://docs.ionos.com/cloud/ai/mcp-server/connect-to-an-ai-client) at docs.ionos.com.

## Tool loading mode

The `IONOS_MCP_LOAD_MODE` environment variable selects how tools are exposed:

- **`eager`** (default): all tools register at startup. Recommended for Claude Code (which defers full schemas client-side via ToolSearch, paying ~1–3k tokens for names only) and the only working mode for clients that ignore `notifications/tools/list_changed` (Claude Desktop, claude.ai connectors, Claude in Chrome, Smithery scanner).

- **`lazy`**: Compute and Object Storage register only on demand. Two sentinel tools (`ionos_load_compute_tools`, `ionos_load_objectstorage_tools`) appear at startup; calling either registers the full product set and emits `notifications/tools/list_changed`. Use only if your MCP client honours that notification AND lacks client-side schema deferral — otherwise eager mode is cheaper.

Parsing is case-insensitive. Unknown values fall back to `eager`.

```json
{
  "mcpServers": {
    "ionoscloud": {
      "command": "/path/to/ionoscloud-mcp",
      "env": {
        "IONOS_TOKEN": "your-api-token",
        "IONOS_MCP_LOAD_MODE": "lazy"
      }
    }
  }
}
```

**Tool-count limits:** Windsurf caps connected MCP servers at 100 tools combined. With the default eager mode the server exposes 112 tools and exceeds that limit. Use lazy loading on Windsurf. For more information, see [Selective Tool Loading](https://docs.ionos.com/cloud/ai/mcp-server/configuration/selective-tool-loading).

## Demo

In Claude Desktop or any other supported client, after configuring the server, try one of these prompts. They cover the kinds of multi-step inspection workflows that are tedious to write as scripts but easy in natural language:

- **Cost audit:** *"Audit my IONOS CLOUD account, find the top 5 cost-inducing resources this month, and suggest cost-efficiency tips."*
- **Security sweep:** *"List every bucket whose public access block is off or whose policy is public — flag anything that looks unintentional."*
- **Audit trail:** *"Show me every failed API request on my contract in the last 30 days, grouped by user."*
- **Forgotten resources:** *"Find unattached volumes, unused IP blocks, and stopped servers across all my data centers."*
- **DNS sanity check:** *"List all zones on my account and flag any without DNSSEC enabled or with records pointing to IPs I no longer own."*
- **Certificate expiry:** *"Which certificates on my account expire in the next 60 days?"*
- **Traffic spike investigation:** *"My last invoice was higher than usual — show me daily traffic and utilization for the previous billing period and tell me what changed."*
- **Onboarding tour:** *"Walk me through what I have running on IONOS CLOUD — datacenters, servers, storage, DNS — like you're explaining it to a new teammate."*

Each prompt chains multiple `list_*` and `get_*` calls and produces a summary you can paste into a ticket, dashboard, or doc. For end-to-end walkthroughs:

* [Run a security posture audit on your IONOS CLOUD Object Storage buckets](https://docs.ionos.com/cloud/tutorials/ai/mcp-server/object-storage-security-audit)
* [Generate a FOCUS-compliant cost report](https://docs.ionos.com/cloud/tutorials/ai/mcp-server/focus-billing-finops)

## Development

### Testing the MCP protocol locally

You can test the server's MCP protocol implementation using stdin/stdout:

```bash
# Initialize and list tools
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  sleep 1
} | ./ionoscloud-mcp

# Call a tool (requires a valid IONOS_TOKEN)
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_datacenters","arguments":{}}}'
  sleep 1
} | ./ionoscloud-mcp
```

### Building from source

```bash
make build
# or
go build -o ionoscloud-mcp .
```

Run `make` with no arguments to see the available targets.

## Related projects

* [IONOS CLOUD MCP Server product docs](https://docs.ionos.com/cloud/ai/mcp-server) — full product documentation
* [IONOS CLOUD AI Model Hub](https://docs.ionos.com/cloud/ai/ai-model-hub) — open-weight LLMs hosted in Germany; pair with this server for a fully EU-sovereign AI loop
* [IONOS CLOUD Documentation MCP](https://docs.ionos.com/cloud/~gitbook/mcp) — a free public MCP server exposing the IONOS docs site for AI assistants

## Contributing

Issues and pull requests are welcome. For development setup, code style, and testing instructions, see [CONTRIBUTING.md](CONTRIBUTING.md). For questions and discussion, use [GitHub Discussions](https://github.com/ionos-cloud/ionoscloud-mcp/discussions).

## Security

If you believe you have found a security vulnerability, **please do not open a public issue**. Report it privately via GitHub's [private vulnerability reporting](https://github.com/ionos-cloud/ionoscloud-mcp/security/advisories/new) or by email to `sdk-tooling@ionos.com`. Full policy: [SECURITY.md](SECURITY.md).

## Changelog

Notable changes per release are tracked in [CHANGELOG.md](CHANGELOG.md). For the artefacts published with each tag (Linux/macOS/Windows binaries, multi-arch OCI images), see the [GitHub Releases](https://github.com/ionos-cloud/ionoscloud-mcp/releases) page.

## API documentation

For more information about the IONOS CLOUD API:

* [IONOS CLOUD API Documentation](https://api.ionos.com/docs/)
* [API specifications](https://github.com/ionos-cloud/rest-api/tree/main/public)
* [SDK documentation](https://github.com/ionos-cloud/sdk-go-bundle)

## License

Apache License 2.0 — see [LICENSE](LICENSE).
