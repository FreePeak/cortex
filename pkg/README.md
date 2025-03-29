# MCP Server SDK for Go

The Model Context Protocol (MCP) Server SDK for Go provides a simple way to create MCP-compliant servers in Go. This SDK allows you to:

- Create MCP servers with custom tools
- Handle tool calls with your own business logic
- Serve MCP over standard I/O

## Usage

### Creating a simple MCP server

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/pkg/server"
	"github.com/FreePeak/golang-mcp-server-sdk/pkg/tools"
)

func main() {
	// Create the server
	mcpServer := server.NewMCPServer("My MCP Server", "1.0.0")

	// Create a tool
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Add the tool to the server with a handler
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		log.Fatalf("Error adding tool: %v", err)
	}

	// Start the server over stdio
	if err := mcpServer.ServeStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Echo tool handler
func handleEcho(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract the message parameter
	message, ok := request.Parameters["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'message' parameter")
	}

	// Return the echo response
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}, nil
}
```

### Creating Tools

The SDK provides a fluent interface for creating tools and their parameters:

```go
// Create a calculator tool
calculatorTool := tools.NewTool("calculator",
	tools.WithDescription("Performs basic arithmetic operations"),
	tools.WithString("operation",
		tools.Description("The operation to perform (add, subtract, multiply, divide)"),
		tools.Required(),
	),
	tools.WithNumber("a",
		tools.Description("First number"),
		tools.Required(),
	),
	tools.WithNumber("b",
		tools.Description("Second number"),
		tools.Required(),
	),
)
```

### Handling Tool Calls

Tool handlers receive a `ToolCallRequest` and return a result or an error:

```go
func handleCalculator(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract parameters
	operation, ok := request.Parameters["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'operation' parameter")
	}
	
	a, ok := request.Parameters["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'a' parameter")
	}
	
	b, ok := request.Parameters["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'b' parameter")
	}
	
	// Perform the calculation
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
	
	// Return the result
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Result: %v", result),
			},
		},
	}, nil
}
```

## Package Structure

The SDK consists of several packages:

- `pkg/server`: Core server implementation
- `pkg/tools`: Utilities for creating and configuring tools
- `pkg/types`: Common types and interfaces

## Examples

Check out the `examples` directory for complete working examples. 