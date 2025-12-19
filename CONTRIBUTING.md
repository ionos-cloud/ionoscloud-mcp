# Contributing to IONOS Cloud MCP Server

Thank you for your interest in contributing to the IONOS Cloud MCP Server!

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
# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./ionoscloud-mcp

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./ionoscloud-mcp
```

## Adding New Tools

To add a new tool:

1. Add the tool definition in `main.go` in the `registerTools()` function
2. Add the case handler in `ionos.go` in the `executeTool()` function
3. Implement the tool function in `ionos.go`
4. Update the README.md with documentation for the new tool
5. Test the tool using the MCP protocol

## Pull Request Process

1. Ensure your code passes `make check`
2. Update documentation as needed
3. Add examples for new tools
4. Submit a pull request with a clear description of changes

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
