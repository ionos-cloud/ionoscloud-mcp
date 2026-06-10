# Contributing to IONOS CLOUD MCP Server

Thank you for your interest in contributing to the IONOS CLOUD MCP Server!

## Development Setup

1. Fork and clone the repository
2. Install [Go 1.25 or higher](https://go.dev/dl/) (see `go.mod` for the exact minimum)
3. Install dependencies: `make deps`
4. Build the project: `make build`

Run `make` with no arguments to list all available targets.

## Code Style

- Format code: `make fmt`
- Run `go vet`: `make vet`
- Run both: `make check`
- Run the full linter: `make lint` (or `make lintfix` to auto-fix) — requires [golangci-lint](https://golangci-lint.run/)
- Check for known vulnerabilities: `make vuln` — requires [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)

## Testing

Run the unit tests:

```bash
make test
```

You can also exercise the MCP protocol end to end over stdio:

```bash
# Initialize and list tools
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  sleep 1
} | ./ionoscloud-mcp
```

Calling tools requires a valid `IONOS_TOKEN` in the environment (the server starts without one, but API calls will return 401).

## Adding New Tools

To add a tool to an existing product:

1. Define an input struct in `tools/inputs.go` with `json` and `jsonschema` tags (non-pointer fields are automatically required)
2. Add a `mcp.AddTool()` call in the appropriate resource file under `tools/<product>/` (e.g., `tools/compute/server.go`)
3. If it's a new resource, create a new file and register its function in `tools/<product>/register.go`
4. Document the tool in `docs/<product>/` (e.g., `docs/compute/server.md`) and add it to the index in `docs/README.md`
5. Update the tool counts in the `README.md` products table
6. Test the tool over the MCP protocol

To add a new product:

1. Create a new sub-package under `tools/<product>/` with a `RegisterAll()` function
2. Add the SDK import and client initialization to `main.go`
3. Call `RegisterAll()` from `main()` — directly for eagerly loaded products, or via a loader function in `tools/loader/loader.go` if the product should be lazy-loadable (see `IONOS_MCP_LOAD_MODE` in the README)
4. Add docs under `docs/<product>/` and list them in `docs/README.md`

All tools must be **read-only** (`list_*`, `get_*`, `head_*`). This server does not create, modify, or delete resources by design.

## Pull Request Process

1. Ensure your code passes `make check`, `make lint`, and `make test`
2. Update documentation as needed
3. Add examples for new tools
4. Add an entry to the **Unreleased** section of `CHANGELOG.md`
5. Use conventional commit messages (`feat:`, `fix:`, `chore:`, `doc:`, `test:`)
6. Submit a pull request with a clear description of changes

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
