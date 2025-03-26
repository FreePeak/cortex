# Golang MCP Server SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/FreePeak/golang-mcp-server-sdk.svg)](https://pkg.go.dev/github.com/FreePeak/golang-mcp-server-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/FreePeak/golang-mcp-server-sdk)](https://goreportcard.com/report/github.com/FreePeak/golang-mcp-server-sdk)

A Golang implementation of the Model Context Protocol (MCP) server SDK. This project follows Clean Architecture principles and provides a framework for building MCP-compatible servers.

## Features

- Full implementation of the MCP protocol
- Clean Architecture design for maintainability and testability
- Support for resources, tools, and prompts
- Server-Sent Events (SSE) for real-time notifications
- Thread-safe repositories
- Example server implementation

## Requirements

- Go 1.23 or later

## Installation

```bash
go get github.com/FreePeak/golang-mcp-server-sdk
```

## Project Structure

The project follows Clean Architecture principles, with the following layers:

- **Domain**: Core business entities and interfaces
- **Use Cases**: Application business logic
- **Infrastructure**: Implementation details (repositories, notification systems)
- **Interfaces**: HTTP handlers and JSON-RPC implementation

```
├── cmd
│   └── example         # Example MCP server application
├── internal
│   ├── domain          # Core business entities and interfaces
│   ├── usecases        # Application business logic
│   ├── interfaces      # HTTP/JSON-RPC adapters
│   └── infrastructure  # Implementation details
├── pkg
│   └── utils           # Shared utilities
└── mcp                 # Public SDK types
```

## Quick Start

Run the example server:

```bash
make example
```

This will start an MCP server on port 8080 with sample resources, tools, and prompts.

## Available Make Commands

- `make build`: Build the example server binary
- `make example`: Run the example server
- `make test`: Run all tests
- `make coverage`: Generate test coverage report
- `make lint`: Run the linter
- `make clean`: Clean build artifacts
- `make deps`: Update dependencies

## Usage

### Creating a Custom MCP Server

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

// Create a custom context function that adds a timestamp
func withTimestamp(ctx context.Context) context.Context {
	// Log when the context function is called
	log.Printf("Context function called with context: %v", ctx)
	return context.WithValue(ctx, "timestamp", fmt.Sprintf("%d", time.Now().Unix()))
}

func main() {
	// Configure logger
	logger := log.New(os.Stderr, "[ECHO-STDIO] ", log.LstdFlags)
	logger.Println("Starting Echo Stdio Server...")

	// Create the echo tool definition with proper inputSchema
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

	// Create repositories
	logger.Println("Creating tool repository...")
	toolRepo := server.NewInMemoryToolRepository()
	ctx := context.Background()
	if err := toolRepo.AddTool(ctx, echoTool); err != nil {
		logger.Fatalf("Failed to add echo tool: %v", err)
	}

	// List tools to verify the echo tool was added
	tools, err := toolRepo.ListTools(ctx)
	if err != nil {
		logger.Fatalf("Failed to list tools: %v", err)
	}
	logger.Printf("Repository has %d tools:", len(tools))
	for _, tool := range tools {
		logger.Printf("- Tool: %s - %s", tool.Name, tool.Description)
	}

	// Create server service with the echo tool
	logger.Println("Creating server service...")
	serviceConfig := usecases.ServerConfig{
		Name:         "Echo Stdio Server",
		Version:      "1.0.0",
		Instructions: "This is a simple echo server that echoes back messages sent to it.",
		ToolRepo:     toolRepo,
	}
	service := usecases.NewServerService(serviceConfig)

	// Create MCP server with a dummy address since we won't be using the HTTP server
	logger.Println("Creating MCP server...")
	mcpServer := rest.NewMCPServer(service, ":0")

	// Start the stdio server with our custom context function
	logger.Println("Server ready. You can now send JSON-RPC requests via stdin.")
	err = stdio.ServeStdio(
		mcpServer,
		stdio.WithErrorLogger(logger),
		stdio.WithStdioContextFunc(withTimestamp),
	)

	if err != nil {
		logger.Printf("Error serving stdio: %v", err)
		os.Exit(1)
	}
}

```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 

## 📧 Support & Contact

- For questions or issues, email [mnhatlinh.doan@gmail.com](mailto:mnhatlinh.doan@gmail.com)
- Open an issue directly: [Issue Tracker](https://github.com/FreePeak/db-mcp-server/issues)
- If Golang MCP Server SDK helps your work, please consider supporting:

<p align="">
<a href="https://www.buymeacoffee.com/linhdmn">
<img src="https://img.buymeacoffee.com/button-api/?text=Support DB MCP Server&emoji=☕&slug=linhdmn&button_colour=FFDD00&font_colour=000000&font_family=Cookie&outline_colour=000000&coffee_colour=ffffff" 
alt="Buy Me A Coffee"/>
</a>
</p>

