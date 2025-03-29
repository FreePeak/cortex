# MCP Server SDK Examples

This directory contains example code demonstrating how to use the MCP Server SDK.

## Examples Overview

### 1. Echo Server (`echo_server.go`)

A simple example demonstrating how to create an MCP server with an echo tool.

```bash
# Run the example
go run examples/echo_server.go
```

Test with:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","parameters":{"message":"Hello, World!"}}}' | go run examples/echo_server.go
```

### 2. Calculator Server (`calculator/`)

A more complex example showing how to:
- Create a server that can run in either HTTP or stdio mode
- Use command-line flags to configure server behavior
- Handle parameter validation and type conversion
- Implement graceful shutdown for HTTP servers

```bash
# Run in HTTP mode
go run examples/calculator/main.go --mode http

# Run in stdio mode
go run examples/calculator/main.go --mode stdio
```

Test with:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"calculator","parameters":{"operation":"add","a":5,"b":3}}}' | go run examples/calculator/main.go --mode stdio
```

## Running the Examples

All examples can be compiled and run directly:

```bash
# Compile examples
go build -o bin/echo-server examples/echo_server.go
go build -o bin/calculator-server examples/calculator/main.go

# Run examples
./bin/echo-server
./bin/calculator-server --mode http
```

## Using the Examples as Templates

These examples can serve as templates for your own MCP servers:

1. Start with the `echo_server.go` for a minimal implementation
2. Use the `calculator` example for more advanced features

## Documentation

For more information about the MCP Server SDK, see the [pkg/README.md](../pkg/README.md) file. 