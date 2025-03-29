# Golang MCP Server SDK 
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/FreePeak/golang-mcp-server-sdk)](https://goreportcard.com/report/github.com/FreePeak/golang-mcp-server-sdk)
[![Go Reference](https://pkg.go.dev/badge/github.com/FreePeak/golang-mcp-server-sdk.svg)](https://pkg.go.dev/github.com/FreePeak/golang-mcp-server-sdk)
[![Build Status](https://github.com/FreePeak/golang-mcp-server-sdk/actions/workflows/go.yml/badge.svg)](https://github.com/FreePeak/golang-mcp-server-sdk/actions/workflows/go.yml)
[![Contributors](https://img.shields.io/github/contributors/FreePeak/golang-mcp-server-sdk)](https://github.com/FreePeak/golang-mcp-server-sdk/graphs/contributors)
## Table of Contents
- [Overview](#overview)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [What is MCP?](#what-is-mcp)
- [Core Concepts](#core-concepts)
  - [Server](#server)
  - [Resources](#resources)
  - [Tools](#tools)
  - [Prompts](#prompts)
- [Running Your Server](#running-your-server)
  - [stdio](#stdio)
  - [HTTP with SSE](#http-with-sse)
  - [Multi-Protocol](#multi-protocol)
  - [Testing and Debugging](#testing-and-debugging)
- [Examples](#examples)
  - [Echo Server](#echo-server)
- [Advanced Usage](#advanced-usage)
  - [Builder Pattern](#builder-pattern)
  - [Clean Architecture](#clean-architecture)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Overview

The Model Context Protocol allows applications to provide context for LLMs in a standardized way, separating the concerns of providing context from the actual LLM interaction. This Golang SDK implements the full MCP specification, making it easy to:

- Build MCP servers that expose resources, prompts and tools
- Use standard transports like stdio and Server-Sent Events (SSE)
- Handle all MCP protocol messages and lifecycle events
- Follow Go best practices and clean architecture principles

> **Note:** This SDK is always updated to align with the latest MCP specification from [spec.modelcontextprotocol.io/latest](https://spec.modelcontextprotocol.io/latest)

## Installation

```bash
go get github.com/FreePeak/golang-mcp-server-sdk
```

## Quickstart

Let's create a simple MCP server that exposes an echo tool:

```go
package main

import (
	"context"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
)

func main() {
	// Configure logger
	logger, err := logging.New(logging.Config{
		Level:       logging.InfoLevel,
		Development: true,
		OutputPaths: []string{"stderr"},
		InitialFields: logging.Fields{
			"component": "echo-stdio",
		},
	})
	if err != nil {
		os.Exit(1)
	}

	// Create the echo tool definition
	echoTool := &domain.Tool{
		Name:        "echo_golang_mcp_server_stdio",
		Description: "Echoes back the input message",
		Parameters: []domain.ToolParameter{
			{
				Name:        "message",
				Description: "The message to echo back",
				Type:        "string",
				Required:    true,
			},
		},
	}

	// Use the server builder pattern
	ctx := context.Background()
	serverBuilder := builder.NewServerBuilder().
		WithName("Echo Stdio Server").
		WithVersion("1.0.0").
		WithInstructions("This is a simple echo server that echoes back messages sent to it.").
		AddTool(ctx, echoTool)

	// Start the stdio server
	logger.Info("Server ready. You can now send JSON-RPC requests via stdin.")
	err = serverBuilder.ServeStdio(
		stdio.WithLogger(logger),
	)

	if err != nil {
		logger.Error("Error serving stdio", logging.Fields{"error": err})
		os.Exit(1)
	}
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

The MCP Server is your core interface to the MCP protocol. It handles connection management, protocol compliance, and message routing. With our builder pattern, you can easily configure your server:

```go
serverBuilder := builder.NewServerBuilder().
    WithName("My App").
    WithVersion("1.0.0").
    WithInstructions("Custom instructions for the LLM about how to interact with your server")
```

### Resources

Resources are how you expose data to LLMs. They're similar to GET endpoints in a REST API - they provide data but shouldn't perform significant computation or have side effects:

```go
// Create a resource
resource := &domain.Resource{
    URI:         "sample://hello-world",
    Name:        "Hello World Resource",
    Description: "A sample resource for demonstration purposes",
    MIMEType:    "text/plain",
}

// Add it to the server builder
serverBuilder.AddResource(ctx, resource)
```

### Tools

Tools let LLMs take actions through your server. Unlike resources, tools are expected to perform computation and have side effects:

```go
// Define a calculator tool
calculatorTool := &domain.Tool{
    Name:        "calculate",
    Description: "Performs basic arithmetic",
    Parameters: []domain.ToolParameter{
        {
            Name:        "operation",
            Description: "The operation to perform (add, subtract, multiply, divide)",
            Type:        "string",
            Required:    true,
        },
        {
            Name:        "a", 
            Description: "First number",
            Type:        "number",
            Required:    true,
        },
        {
            Name:        "b",
            Description: "Second number",
            Type:        "number",
            Required:    true,
        },
    },
}

// Add tool to server builder
serverBuilder.AddTool(ctx, calculatorTool)

// Register tool handler after building the service
service := serverBuilder.BuildService()
service.RegisterToolHandler("calculate", func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    operation := params["operation"].(string)
    a := params["a"].(float64)
    b := params["b"].(float64)
    
    var result float64
    switch operation {
    case "add":
        result = a + b
    case "subtract":
        result = a - b
    case "multiply":
        result = a * b
    case "divide":
        result = a / b
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
    
    return result, nil
})
```

### Prompts

Prompts are reusable templates that help LLMs interact with your server effectively:

```go
// Create a prompt for code review
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

// Add prompt to server builder
serverBuilder.AddPrompt(ctx, codeReviewPrompt)
```

## Running Your Server

MCP servers in Go can be connected to different transports depending on your use case:

### stdio

For command-line tools and direct integrations:

```go
// Configure and start a stdio server using the builder
err := serverBuilder.ServeStdio(
    stdio.WithLogger(logger),
    stdio.WithStdioContextFunc(customContextFunction),
)
```

### HTTP with SSE

For web applications, you can use Server-Sent Events (SSE) for real-time communication:

```go
// Build an MCP server with SSE support
mcpServer := serverBuilder.BuildMCPServer()

// Start the server
if err := mcpServer.Start(); err != nil {
    logger.Fatal("Server failed to start", logging.Fields{"error": err})
}

// For graceful shutdown
if err := mcpServer.Stop(ctx); err != nil {
    logger.Fatal("Server forced to shutdown", logging.Fields{"error": err})
}
```

### Multi-Protocol

You can also run multiple protocol servers simultaneously:

```go
// Create a multi-protocol server
serverBuilder := builder.NewServerBuilder().
    WithName("Multi-Protocol Server").
    WithVersion("1.0.0").
    WithAddress(":8080").  // For HTTP/SSE server
    AddTool(ctx, echoTool)

// For HTTP mode
mcpServer := serverBuilder.BuildMCPServer()
go startHTTPServer(mcpServer, logger)

// For StdIO mode (in the same application)
err := serverBuilder.ServeStdio(stdio.WithErrorLogger(logger))
```

### Testing and Debugging

For testing your MCP server, you can use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector).

## Examples

Check out the `cmd` directory for complete example servers:

### Echo Server

A simple echo server with SSE support:

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
)

func main() {
	// Configure logger
	logger, err := logging.New(logging.Config{
		Level:       logging.InfoLevel,
		Development: true,
		OutputPaths: []string{"stdout"},
		InitialFields: logging.Fields{
			"component": "echo-sse",
		},
	})
	if err != nil {
		os.Exit(1)
	}

	// Create the echo tool
	echoTool := &domain.Tool{
		Name:        "mcp_golang_mcp_server_sse_echo",
		Description: "Echoes back the input message",
		Parameters: []domain.ToolParameter{
			{
				Name:        "message",
				Description: "The message to echo back",
				Type:        "string",
				Required:    true,
			},
		},
	}

	// Create a sample resource
	sampleResource := &domain.Resource{
		URI:         "sample://hello-world",
		Name:        "Hello World Resource",
		Description: "A sample resource for demonstration purposes",
		MIMEType:    "text/plain",
	}

	// Use the builder pattern to create the server
	ctx := context.Background()
	serverBuilder := builder.NewServerBuilder().
		WithName("Echo SSE Server").
		WithVersion("1.0.0").
		WithInstructions("This is a simple echo server that echoes back messages sent to it.").
		WithAddress(":8080").
		AddTool(ctx, echoTool).
		AddResource(ctx, sampleResource)

	// Build the MCP server with logger
	service := serverBuilder.BuildService()
	mcpServer := rest.NewMCPServer(service, ":8080", rest.WithLogger(logger))

	// Start server in a goroutine
	go func() {
		if err := mcpServer.Start(); err != nil {
			if err.Error() != "http: Server closed" {
				logger.Fatal("Server failed to start", logging.Fields{"error": err})
			}
		}
	}()

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-shutdown
	logger.Info("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := mcpServer.Stop(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", logging.Fields{"error": err})
	}

	logger.Info("Server stopped gracefully")
}
```

## Advanced Usage

### Builder Pattern

The SDK provides a convenient builder pattern to configure and create your MCP server:

```go
serverBuilder := builder.NewServerBuilder().
    WithName("My Advanced App").
    WithVersion("1.0.0").
    WithInstructions("Detailed instructions for using this MCP server").
    WithAddress(":8080").
    AddTool(ctx, myTool).
    AddResource(ctx, myResource).
    AddPrompt(ctx, myPrompt)

// For a stdio server
serverBuilder.ServeStdio()

// Or for an HTTP/SSE server
mcpServer := serverBuilder.BuildMCPServer()
mcpServer.Start()
```

### Clean Architecture

The SDK follows Clean Architecture principles, separating concerns into layers:

- **Domain**: Core business entities and interfaces
- **Use Cases**: Application business logic
- **Infrastructure**: Implementation details
- **Interfaces**: Adapters for different transport protocols

This design makes it easy to:
- Test components in isolation
- Replace implementations without affecting business logic
- Add new transport protocols without changing core functionality

## Documentation

- [Model Context Protocol documentation](https://modelcontextprotocol.io)
- [MCP Specification](https://spec.modelcontextprotocol.io/latest)
- [Example Servers](https://github.com/FreePeak/golang-mcp-server-sdk/tree/main/cmd)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📧 Support & Contact

- For questions or issues, email [mnhatlinh.doan@gmail.com](mailto:mnhatlinh.doan@gmail.com)
- Open an issue directly: [Issue Tracker](https://github.com/FreePeak/golang-mcp-server-sdk/issues)
- If Golang MCP Server SDK helps your work, please consider supporting:

<p align="">
<a href="https://www.buymeacoffee.com/linhdmn">
<img src="https://img.buymeacoffee.com/button-api/?text=Support MCP Server SDK&emoji=☕&slug=linhdmn&button_colour=FFDD00&font_colour=000000&font_family=Cookie&outline_colour=000000&coffee_colour=ffffff" 
alt="Buy Me A Coffee"/>
</a>
</p>

