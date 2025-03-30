# Cortex Examples

This directory contains example applications that demonstrate how to use the Cortex MCP server platform in various scenarios.

## Server Examples

These examples showcase different server configurations and communication protocols:

- **stdio-server**: A simple stdio-based MCP server example
- **sse-server**: An HTTP/SSE-based MCP server example
- **multi-protocol**: A server that supports both stdio and HTTP protocols

## Tool Provider Examples

The `providers` directory contains example tool providers that can be used with the Cortex platform:

- **weather**: Weather forecast tool provider
- **database**: Simple key-value store tool provider

## Running the Examples

### Stdio Server

```bash
go run examples/stdio-server/main.go
```

This will start a stdio server that accepts JSON-RPC requests from standard input.

### SSE Server

```bash
go run examples/sse-server/main.go
```

This will start an HTTP server on port 8080 that accepts MCP requests via Server-Sent Events (SSE).

### Multi-Protocol Server

```bash
# Run with stdio protocol
go run examples/multi-protocol/main.go -protocol stdio

# Run with HTTP protocol
go run examples/multi-protocol/main.go -protocol http -address localhost:8080
```

This example shows how to create a server that can switch between stdio and HTTP protocols. 