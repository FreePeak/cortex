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
    "github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
    "github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
    "github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
    "github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

func main() {
    // Create repositories
    resourceRepo := server.NewInMemoryResourceRepository()
    toolRepo := server.NewInMemoryToolRepository()
    promptRepo := server.NewInMemoryPromptRepository()
    sessionRepo := server.NewInMemorySessionRepository()
    notifier := server.NewNotificationSender("2.0")

    // Create service
    service := usecases.NewServerService(usecases.ServerConfig{
        Name:               "My MCP Server",
        Version:            "1.0.0",
        Instructions:       "Server instructions for LLMs",
        ResourceRepo:       resourceRepo,
        ToolRepo:           toolRepo,
        PromptRepo:         promptRepo,
        SessionRepo:        sessionRepo,
        NotificationSender: notifier,
    })

    // Create and start HTTP server
    mcpServer := rest.NewMCPServer(service, ":8080")
    mcpServer.Start()
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 