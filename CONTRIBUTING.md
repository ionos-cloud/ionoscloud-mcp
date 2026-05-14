# Contributing to IONOS CLOUD MCP Server

Thank you for your interest in contributing to the IONOS CLOUD MCP Server!

## Development Setup

1. Fork and clone the repository
2. Install Go 1.20 or higher
3. Install dependencies: `make deps`
4. Build the project: `make build`

## Code Style

- Follow standard Go formatting: `make fmt`
- Run linter checks: `make vet`
- Run both: `make check`

## Testing

Test your changes using the MCP protocol:

```bash
# Initialize and list tools
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  sleep 1
} | ./ionoscloud-mcp
```

## Adding New Tools

To add a new tool:

1. Define an input struct in `tools/inputs.go` with `json` and `jsonschema` tags (non-pointer fields are automatically required)
2. Add a `mcp.AddTool()` call in the appropriate resource file under `tools/<product>/` (e.g., `tools/compute/server.go`)
3. Register the new resource file's function in `tools/<product>/register.go` if it's a new resource
4. Add the tool to the appropriate `docs/<product>/` resource file (e.g., `docs/compute/server.md`)
5. Update the tools table in `README.md`
6. Test the tool using the MCP protocol

## Pull Request Process

1. Ensure your code passes `make check`
2. Update documentation as needed
3. Add examples for new tools
4. Add an entry to CHANGELOG.md
5. Submit a pull request with a clear description of changes

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
