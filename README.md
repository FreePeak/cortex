# Golang MCP Server SDK ![Go Reference](https://pkg.go.dev/badge/github.com/FreePeak/golang-mcp-server-sdk.svg) ![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)

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
  - [WebSockets](#websockets)
  - [Multi-Protocol](#multi-protocol)
  - [Testing and Debugging](#testing-and-debugging)
- [Examples](#examples)
  - [Echo Server](#echo-server)
- [Advanced Usage](#advanced-usage)
  - [Low-Level Server](#low-level-server)
  - [Clean Architecture](#clean-architecture)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Overview

The Model Context Protocol allows applications to provide context for LLMs in a standardized way, separating the concerns of providing context from the actual LLM interaction. This Golang SDK implements the full MCP specification, making it easy to:

- Build MCP servers that expose resources, prompts and tools
- Use standard transports like stdio, SSE, and WebSockets
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
	"log"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

func main() {
	// Configure logger
	logger := log.New(os.Stderr, "[ECHO-SERVER] ", log.LstdFlags)
	
	// Create the echo tool
	echoTool := &domain.Tool{
		Name:        "echo",
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

	// Create tool repository
	toolRepo := server.NewInMemoryToolRepository()
	ctx := context.Background()
	toolRepo.AddTool(ctx, echoTool)

	// Create server service
	serviceConfig := usecases.ServerConfig{
		Name:     "Echo Server",
		Version:  "1.0.0",
		ToolRepo: toolRepo,
	}
	service := usecases.NewServerService(serviceConfig)

	// Create MCP server
	mcpServer := stdio.NewStdioServer(service, logger)

	// Start the stdio server
	if err := mcpServer.Serve(); err != nil {
		logger.Fatalf("Error serving: %v", err)
	}
}
```

## What is MCP?

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io) is a standardized protocol that allows applications to provide context for LLMs in a secure and efficient manner. It separates the concerns of providing context and tools from the actual LLM interaction. MCP servers can:

- Expose data through **Resources** (read-only data endpoints)
- Provide functionality through **Tools** (executable functions)
- Define interaction patterns through **Prompts** (reusable templates)
- Support various transport methods (stdio, HTTP/SSE, WebSockets)

## Core Concepts

### Server

The MCP Server is your core interface to the MCP protocol. It handles connection management, protocol compliance, and message routing:

```go
serviceConfig := usecases.ServerConfig{
    Name:     "My App",
    Version:  "1.0.0",
    ToolRepo: toolRepo,
}
service := usecases.NewServerService(serviceConfig)
```

### Resources

Resources are how you expose data to LLMs. They're similar to GET endpoints in a REST API - they provide data but shouldn't perform significant computation or have side effects:

```go
// Create resource repository
resourceRepo := server.NewInMemoryResourceRepository()

// Add a static resource
resource := &domain.Resource{
    Name:        "config",
    Description: "Application configuration",
    Uri:         "config://app",
    Content:     "App configuration data goes here",
}
resourceRepo.AddResource(ctx, resource)

// Add resource repository to server config
serviceConfig.ResourceRepo = resourceRepo
```

### Tools

Tools let LLMs take actions through your server. Unlike resources, tools are expected to perform computation and have side effects:

```go
// Define a simple calculator tool
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

// Add tool implementation
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
// Create prompt repository
promptRepo := server.NewInMemoryPromptRepository()

// Add a code review prompt
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
promptRepo.AddPrompt(ctx, codeReviewPrompt)

// Add prompt repository to server config
serviceConfig.PromptRepo = promptRepo
```

## Running Your Server

MCP servers in Go can be connected to different transports depending on your use case:

### stdio

For command-line tools and direct integrations:

```go
import (
    "github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
)

// Create stdio server
stdioServer := stdio.NewStdioServer(service, logger)

// Start the stdio server
if err := stdioServer.Serve(); err != nil {
    logger.Fatalf("Error serving: %v", err)
}
```

### HTTP with SSE

For remote servers, start a web server with a Server-Sent Events (SSE) endpoint:

```go
import (
    "github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
)

// Create HTTP server with SSE support
httpServer := rest.NewMCPServer(service, ":8080")

// Start the HTTP server
if err := httpServer.Serve(); err != nil {
    logger.Fatalf("Error serving: %v", err)
}
```

### WebSockets

For bidirectional communication:

```go
import (
    "github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/ws"
)

// Create WebSocket server
wsServer := ws.NewWebSocketServer(service, ":8081")

// Start the WebSocket server
if err := wsServer.Serve(); err != nil {
    logger.Fatalf("Error serving: %v", err)
}
```

### Multi-Protocol

You can also run multiple protocol servers simultaneously:

```go
import (
    "github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
)

// Create a multi-protocol server builder
serverBuilder := builder.NewServerBuilder(service)

// Add protocols
serverBuilder.WithStdio(logger)
serverBuilder.WithSSE(":8080")
serverBuilder.WithWebSocket(":8081")

// Build and run the server
multiServer, err := serverBuilder.Build()
if err != nil {
    logger.Fatalf("Error building server: %v", err)
}

// Start all servers
if err := multiServer.Serve(); err != nil {
    logger.Fatalf("Error serving: %v", err)
}
```

### Testing and Debugging

For testing your MCP server, you can use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector).

## Examples

Check out the `cmd` directory for complete example servers:

### Echo Server

A simple server demonstrating tools functionality:

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

func main() {
	logger := log.New(os.Stderr, "[ECHO] ", log.LstdFlags)
	
	// Create the echo tool
	echoTool := &domain.Tool{
		Name:        "echo",
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

	// Create tool repository
	toolRepo := server.NewInMemoryToolRepository()
	ctx := context.Background()
	toolRepo.AddTool(ctx, echoTool)

	// Create server service
	serviceConfig := usecases.ServerConfig{
		Name:     "Echo Server",
		Version:  "1.0.0",
		ToolRepo: toolRepo,
	}
	service := usecases.NewServerService(serviceConfig)
	
	// Register tool handler
	service.RegisterToolHandler("echo", func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		message := params["message"].(string)
		return message, nil
	})

	// Create stdio server
	stdioServer := stdio.NewStdioServer(service, logger)

	// Start the stdio server
	if err := stdioServer.Serve(); err != nil {
		logger.Fatalf("Error serving: %v", err)
	}
}
```

## Advanced Usage

### Low-Level Server

For more control, you can use the low-level server interfaces directly:

```go
import (
	"context"
	"encoding/json"
	"log"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/protocol"
)

func handleRequest(ctx context.Context, request []byte) ([]byte, error) {
	var mcpRequest protocol.Request
	if err := json.Unmarshal(request, &mcpRequest); err != nil {
		return nil, err
	}

	// Process the request based on method
	switch mcpRequest.Method {
	case "capabilities":
		response := protocol.CapabilitiesResponse{
			JSONRPC: "2.0",
			ID:      mcpRequest.ID,
			Result: protocol.Capabilities{
				Resources: &protocol.ResourceCapabilities{},
				Tools:     &protocol.ToolCapabilities{},
			},
		}
		return json.Marshal(response)
		
	// Handle other methods similarly
	default:
		return nil, fmt.Errorf("unknown method: %s", mcpRequest.Method)
	}
}
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

