# Example MCP Servers

This directory contains example implementations of MCP servers using the SDK. These examples show how to build MCP-compatible servers for different transport protocols.

## Examples

### 1. Echo StdIO Server (cmd/echo-stdio-server)

A simple MCP server that communicates over standard input/output (stdio). This is useful for command-line tools and direct integration with other processes.

```bash
# Build the server
go build -o bin/echo-stdio-server cmd/echo-stdio-server/main.go

# Run the server
./bin/echo-stdio-server
```

Test with a JSON-RPC request:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo_golang_mcp_server_stdio","parameters":{"message":"Hello, World!"}}}' | ./bin/echo-stdio-server
```

### 2. Echo SSE Server (cmd/echo-sse-server)

An MCP server that communicates over HTTP with Server-Sent Events (SSE). This is useful for web applications.

```bash
# Build the server
go build -o bin/echo-sse-server cmd/echo-sse-server/main.go

# Run the server
./bin/echo-sse-server
```

Connect to the server from Cursor by going to Settings > Extensions > Model Context Protocol and entering 'http://localhost:8080' as the server URL.

### 3. Multi-Protocol Server (cmd/multi-protocol-server)

An MCP server that can run in multiple modes: HTTP, stdio, or both.

```bash
# Build the server
go build -o bin/multi-protocol-server cmd/multi-protocol-server/main.go

# Run in HTTP mode
./bin/multi-protocol-server --mode http

# Run in stdio mode
./bin/multi-protocol-server --mode stdio

# Run in both modes
./bin/multi-protocol-server --mode both
```

## Using the SDK

All examples now use the public SDK interface from the `pkg/` directory. The main components are:

- `pkg/server`: Core server implementation
- `pkg/tools`: Utilities for creating and configuring tools
- `pkg/types`: Common types and interfaces

### Creating a Basic MCP Server

```go
import (
    "github.com/FreePeak/golang-mcp-server-sdk/pkg/server"
    "github.com/FreePeak/golang-mcp-server-sdk/pkg/tools"
)

// Create a server
mcpServer := server.NewMCPServer("My Server", "1.0.0")

// Create a tool with the fluent API
myTool := tools.NewTool("my_tool",
    tools.WithDescription("My tool description"),
    tools.WithString("param1",
        tools.Description("Parameter description"),
        tools.Required(),
    ),
)

// Add the tool with a handler
ctx := context.Background()
err := mcpServer.AddTool(ctx, myTool, handleMyTool)

// Start the server
mcpServer.ServeStdio()
```

### Tool Handler Function

```go
func handleMyTool(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
    // Extract parameters
    param1, ok := request.Parameters["param1"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid parameter")
    }
    
    // Process and return result
    return map[string]interface{}{
        "content": []map[string]interface{}{
            {
                "type": "text",
                "text": fmt.Sprintf("Processed: %s", param1),
            },
        },
    }, nil
}
```

For more information, see the main SDK documentation in the `pkg/README.md` file. 