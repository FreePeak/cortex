#!/bin/bash

# Build the echo server
go build -o bin/echo-stdio-server cmd/echo-stdio-server/main.go

# Test tools/call command
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo_golang_mcp_server_stdio","parameters":{"message":"Hello, MCP Server!"}}}' | ./bin/echo-stdio-server

echo "Test completed!"
