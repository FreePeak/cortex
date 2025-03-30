#!/bin/bash

# Build the echo server
go build -o bin/echo-stdio-server cmd/echo-stdio-server/main.go

# Test calling the tool with the platform-prefixed name
echo "Testing platform-prefixed tool 'cortex_echo'..."
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"cortex_echo","parameters":{"message":"Hello, MCP Server!"}}}' | \
CORTEX_DEBUG=1 ./bin/echo-stdio-server

echo "Test completed!"
