<h1 align="center">
   <img alt="main logo" src="logo.svg" width="150"/>
   <br/>
   Cortex
</h1>
<h4 align="center">Build MCP Servers Declaratively in Golang</h4>

<p align="center">
	<a href="https://pkg.go.dev/github.com/FreePeak/cortex"><img src="https://pkg.go.dev/badge/github.com/FreePeak/cortex.svg" alt="Go Reference"></a>
	<a href="https://goreportcard.com/report/github.com/FreePeak/cortex"><img src="https://goreportcard.com/badge/github.com/FreePeak/cortex" alt="Go Report Card"></a>
	<a href="https://github.com/FreePeak/cortex/actions/workflows/go.yml"><img src="https://github.com/FreePeak/cortex/actions/workflows/go.yml/badge.svg" alt="Go Workflow"></a>
	<a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
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
  - [Providers](#providers)
  - [Resources](#resources)
  - [Prompts](#prompts)
- [Running Your Server](#running-your-server)
  - [STDIO](#stdio)
  - [HTTP with SSE](#http-with-sse)
  - [Multi-Protocol](#multi-protocol)
  - [Testing and Debugging](#testing-and-debugging)
- [Examples](#examples)
  - [Basic Examples](#basic-examples)
  - [Advanced Examples](#advanced-examples)
  - [Plugin System](#plugin-system)
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
	// Create a logger that writes to stderr instead of stdout
	// This is critical for STDIO servers as stdout must only contain JSON-RPC messages
	logger := log.New(os.Stderr, "[cortex] ", log.LstdFlags)

	// Create the server
	mcpServer := server.NewMCPServer("Echo Server Example", "1.0.0", logger)

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
		logger.Fatalf("Error adding tool: %v", err)
	}

	// Write server status to stderr instead of stdout to maintain clean JSON protocol
	fmt.Fprintf(os.Stderr, "Starting Echo Server...\n")
	fmt.Fprintf(os.Stderr, "Send JSON-RPC messages via stdin to interact with the server.\n")
	fmt.Fprintf(os.Stderr, `Try: {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","parameters":{"message":"Hello, World!"}}}\n`)

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
// Create a new MCP server with logger
mcpServer := server.NewMCPServer("My App", "1.0.0", logger)
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

### Providers

Providers allow you to group related tools and resources into a single package that can be easily registered with a server:

```go
// Create a weather provider
weatherProvider, err := weather.NewWeatherProvider(logger)
if err != nil {
    logger.Fatalf("Failed to create weather provider: %v", err)
}

// Register the provider with the server
err = mcpServer.RegisterProvider(ctx, weatherProvider)
if err != nil {
    logger.Fatalf("Failed to register weather provider: %v", err)
}
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

### STDIO

For command-line tools and direct integrations:

```go
// Start a stdio server
if err := mcpServer.ServeStdio(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

IMPORTANT: When using STDIO, all logs must be directed to stderr to maintain the clean JSON-RPC protocol on stdout:

```go
// Create a logger that writes to stderr
logger := log.New(os.Stderr, "[cortex] ", log.LstdFlags)

// All debug/status messages should use stderr
fmt.Fprintf(os.Stderr, "Server starting...\n")
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

You can also run multiple protocol servers simultaneously by using goroutines:

```go
// Start an HTTP server
go func() {
    if err := mcpServer.ServeHTTP(); err != nil {
        log.Fatalf("HTTP server error: %v", err)
    }
}()

// Start a STDIO server
go func() {
    if err := mcpServer.ServeStdio(); err != nil {
        log.Fatalf("STDIO server error: %v", err)
    }
}()

// Wait for shutdown signal
stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
<-stop
```

### Testing and Debugging

For testing and debugging, the Cortex framework provides several utilities:

```go
// You can use the test-call.sh script to send test requests to your STDIO server
// For example:
// ./test-call.sh echo '{"message":"Hello, World!"}'
```

## Examples

### Basic Examples

The repository includes several basic examples in the `examples` directory:

- **STDIO Server**: A simple MCP server that communicates via STDIO (`examples/stdio-server`)
- **SSE Server**: A server that uses HTTP with Server-Sent Events for communication (`examples/sse-server`)
- **Multi-Protocol**: A server that can run on multiple protocols simultaneously (`examples/multi-protocol`)

### Advanced Examples

The examples directory also includes more advanced use cases:

- **Providers**: Examples of how to create and use providers to organize related tools (`examples/providers`)
  - **Weather Provider**: Demonstrates how to create a provider for weather-related tools
  - **Database Provider**: Shows how to create a provider for database operations

### Plugin System

Cortex includes a plugin system for extending server capabilities:

```go
// Create a new provider based on the BaseProvider
type MyProvider struct {
    *plugin.BaseProvider
}

// Create a new provider instance
func NewMyProvider(logger *log.Logger) (*MyProvider, error) {
    info := plugin.ProviderInfo{
        ID:          "my-provider",
        Name:        "My Provider",
        Version:     "1.0.0",
        Description: "A custom provider for my tools",
        Author:      "Your Name",
        URL:         "https://github.com/yourusername/myrepo",
    }
    
    baseProvider := plugin.NewBaseProvider(info, logger)
    provider := &MyProvider{
        BaseProvider: baseProvider,
    }
    
    // Register tools with the provider
    // ...
    
    return provider, nil
}
```

## Package Structure

The Cortex codebase is organized into several packages:

- `pkg/server`: Core server implementation
- `pkg/tools`: Tool creation and management
- `pkg/plugin`: Plugin system for extending server capabilities
- `pkg/types`: Common types and interfaces
- `pkg/builder`: Builders for creating complex objects

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## Support & Contact

- For questions or issues, email [mnhatlinh.doan@gmail.com](mailto:mnhatlinh.doan@gmail.com)
- Open an issue directly: [Issue Tracker](https://github.com/FreePeak/cortex/issues)
- If Cortex helps your work, please consider supporting:

<p align="">
<a href="https://www.buymeacoffee.com/linhdmn">
<img src="https://img.buymeacoffee.com/button-api/?text=Support Cortex&emoji=☕&slug=linhdmn&button_colour=FFDD00&font_colour=000000&font_family=Cookie&outline_colour=000000&coffee_colour=ffffff" 
alt="Buy Me A Coffee"/>
</a>
</p>