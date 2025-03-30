<h1 align="center">
   <img alt="main logo" src="logo.svg" width="150"/>
   <br/>
   Cortex
</h1>
<h4 align="center">Build MCP Servers Declaratively in Golang</h4>

<p align="center">
	<a href="https://pkg.go.dev/github.com/FreePeak/cortex"><img src="https://pkg.go.dev/badge/github.com/FreePeak/cortex.svg" alt="Go Reference"></a>
	<a href="https://goreportcard.com/report/github.com/FreePeak/cortex"><img src="https://goreportcard.com/badge/github.com/FreePeak/cortex" alt="Go Report Card"></a>
	<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
	<a href="https://github.com/FreePeak/cortex/graphs/contributors"><img src="https://img.shields.io/github/contributors/FreePeak/cortex" alt="Contributors"></a>
</p>

## Table of Contents
- [Overview](#overview)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [What is MCP?](#what-is-mcp)
- [Core Concepts](#core-concepts)
  - [Server](#server)
  - [Tools](#tools)
  - [Resources](#resources)
  - [Prompts](#prompts)
- [Running Your Server](#running-your-server)
  - [stdio](#stdio)
  - [HTTP with SSE](#http-with-sse)
  - [Multi-Protocol](#multi-protocol)
  - [Testing and Debugging](#testing-and-debugging)
- [Examples](#examples)
  - [Echo Server](#echo-server)
  - [Calculator Server](#calculator-server)
- [Package Structure](#package-structure)
- [Contributing](#contributing)
- [License](#license)
- [Support & Contact](#support--contact)

## Overview

The Model Context Protocol allows applications to provide context for LLMs in a standardized way, separating the concerns of providing context from the actual LLM interaction. Cortex implements the full MCP specification, making it easy to:

- Build MCP servers that expose resources and tools
- Use standard transports like stdio and Server-Sent Events (SSE)
- Handle all MCP protocol messages and lifecycle events
- Follow Go best practices and clean architecture principles

> **Note:** Cortex is always updated to align with the latest MCP specification from [spec.modelcontextprotocol.io/latest](https://spec.modelcontextprotocol.io/latest)

## Installation

```bash
go get github.com/FreePeak/cortex
```

## Quickstart

Let's create a simple MCP server that exposes an echo tool:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

func main() {
	// Create the server
	mcpServer := server.NewMCPServer("Echo Server Example", "1.0.0")

	// Create an echo tool
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

	// Start the server
	fmt.Println("Starting Echo Server...")
	fmt.Println("Send JSON-RPC messages via stdin to interact with the server.")
	fmt.Println("Try: {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"echo\",\"parameters\":{\"message\":\"Hello, World!\"}}}")

	// Serve over stdio
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

	// Return the echo response in the format expected by the MCP protocol
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

## What is MCP?

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io) is a standardized protocol that allows applications to provide context for LLMs in a secure and efficient manner. It separates the concerns of providing context and tools from the actual LLM interaction. MCP servers can:

- Expose data through **Resources** (read-only data endpoints)
- Provide functionality through **Tools** (executable functions)
- Define interaction patterns through **Prompts** (reusable templates)
- Support various transport methods (stdio, HTTP/SSE)

## Core Concepts

### Server

The MCP Server is your core interface to the MCP protocol. It handles connection management, protocol compliance, and message routing:

```go
// Create a new MCP server
mcpServer := server.NewMCPServer("My App", "1.0.0")
```

### Tools

Tools let LLMs take actions through your server. Unlike resources, tools are expected to perform computation and have side effects:

```go
// Define a calculator tool
calculatorTool := tools.NewTool("calculator",
    tools.WithDescription("Performs basic math operations"),
    tools.WithString("operation",
        tools.Description("The operation to perform (add, subtract, multiply, divide)"),
        tools.Required(),
    ),
    tools.WithNumber("a", 
        tools.Description("First operand"),
        tools.Required(),
    ),
    tools.WithNumber("b", 
        tools.Description("Second operand"),
        tools.Required(),
    ),
)

// Add the tool to the server with a handler
mcpServer.AddTool(ctx, calculatorTool, handleCalculator)
```

### Platform Tool Naming

When you register a tool with `AddTool()`, it is automatically registered with a platform-prefixed name:

```go
// Original name in your code: "calculator"
calculatorTool := tools.NewTool("calculator", ...)

// After registration, the tool is available as: "cortex_calculator"
mcpServer.AddTool(ctx, calculatorTool, handleCalculator)
```

This means that when calling the tool externally, you must always use the platform-prefixed name:

```go
// Tool must be called using the platform-prefixed name:
// {"jsonrpc":"2.0","method":"tools/call","params":{"name":"cortex_calculator","parameters":{...}}}
```

### Resources

Resources are how you expose data to LLMs. They're similar to GET endpoints in a REST API - they provide data but shouldn't perform significant computation or have side effects:

```go
// Create a resource (Currently using the internal API)
resource := &domain.Resource{
    URI:         "sample://hello-world",
    Name:        "Hello World Resource",
    Description: "A sample resource for demonstration purposes",
    MIMEType:    "text/plain",
}

// Note: Resource support is being updated in the public API
```

### Prompts

Prompts are reusable templates that help LLMs interact with your server effectively:

```go
// Create a prompt (Currently using the internal API)
codeReviewPrompt := &domain.Prompt{
    Name:        "review-code",
    Description: "A prompt for code review",
    Template:    "Please review this code:\n\n{{.code}}",
    Parameters: []domain.PromptParameter{
        {
            Name:        "code",
            Description: "The code to review",
            Type:        "string",
            Required:    true,
        },
    },
}

// Note: Prompt support is being updated in the public API
```

## Running Your Server

MCP servers in Go can be connected to different transports depending on your use case:

### stdio

For command-line tools and direct integrations:

```go
// Start a stdio server
if err := mcpServer.ServeStdio(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

### HTTP with SSE

For web applications, you can use Server-Sent Events (SSE) for real-time communication:

```go
// Configure the HTTP address
mcpServer.SetAddress(":8080")

// Start an HTTP server with SSE support
if err := mcpServer.ServeHTTP(); err != nil {
    log.Fatalf("HTTP server error: %v", err)
}

// For graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := mcpServer.Shutdown(ctx); err != nil {
    log.Fatalf("Server shutdown error: %v", err)
}
```

### Multi-Protocol

You can also run multiple protocol servers simultaneously:

```go
// Configure server for both HTTP and stdio
mcpServer := server.NewMCPServer("Multi-Protocol Server", "1.0.0")
mcpServer.SetAddress(":8080")
mcpServer.AddTool(ctx, echoTool, handleEcho)

// Start HTTP server in a goroutine
go func() {
    if err := mcpServer.ServeHTTP(); err != nil {
        log.Fatalf("HTTP server error: %v", err)
    }
}()

// Start stdio server in the main thread
if err := mcpServer.ServeStdio(); err != nil {
    log.Fatalf("Stdio server error: %v", err)
}
```

### Testing and Debugging

For testing your MCP server, you can use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) or send JSON-RPC messages directly:

```bash
# Test an echo tool with stdio
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","parameters":{"message":"Hello, World!"}}}' | go run your_server.go
```

## Examples

Check out the `examples` directory for complete example servers:

### Echo Server

A simple echo server example is available in `examples/echo_server.go`:

```bash
# Run the example
go run examples/echo_server.go
```

Test with:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","parameters":{"message":"Hello, World!"}}}' | go run examples/echo_server.go
```

### Calculator Server

A more advanced calculator example with both HTTP and stdio modes is available in `examples/calculator/`:

```bash
# Run in HTTP mode
go run examples/calculator/main.go --mode http

# Run in stdio mode
go run examples/calculator/main.go --mode stdio
```

## Package Structure

The library is organized following clean architecture principles:

```
cortex/
├── pkg/                    # Public API (exposed to users)
│   ├── builder/            # Public builder pattern for server construction
│   ├── server/             # Public server implementation
│   ├── tools/              # Utilities for creating MCP tools
│   └── types/              # Shared types and interfaces
├── internal/               # Private implementation details
├── examples/               # Example code snippets and use cases
└── cmd/                    # Example MCP server applications
```

The `pkg/` directory contains all publicly exposed APIs that users of the library should interact with.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📧 Support & Contact

- For questions or issues, email [mnhatlinh.doan@gmail.com](mailto:mnhatlinh.doan@gmail.com)
- Open an issue directly: [Issue Tracker](https://github.com/FreePeak/cortex/issues)
- If Cortex helps your work, please consider supporting:

<p align="">
<a href="https://www.buymeacoffee.com/linhdmn">
<img src="https://img.buymeacoffee.com/button-api/?text=Support Cortex&emoji=☕&slug=linhdmn&button_colour=FFDD00&font_colour=000000&font_family=Cookie&outline_colour=000000&coffee_colour=ffffff" 
alt="Buy Me A Coffee"/>
</a>
</p>

