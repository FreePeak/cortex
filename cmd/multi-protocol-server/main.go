package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
)

const (
	serverName      = "Multi-Protocol MCP Server"
	serverVersion   = "1.0.0"
	defaultAddr     = ":8080"
	shutdownTimeout = 10 * time.Second

	serverInstructions = `
This is a multi-protocol MCP server example that can run in HTTP, SSE, or StdIO mode.
It demonstrates how to use the builder pattern to create different server types
with a shared configuration.

Available tool:
- mcp_golang_mcp_server_ws_mcp_golang_mcp_server_sse_echo: Echoes back the input message
`
)

func main() {
	// Define command line flags for server mode
	mode := flag.String("mode", "http", "Server mode: http, stdio, or both")
	addr := flag.String("addr", defaultAddr, "HTTP server address")
	flag.Parse()

	// Setup logger
	logger := log.New(os.Stdout, "[MCP-SERVER] ", log.LstdFlags|log.Lshortfile)
	logger.Printf("Starting %s v%s in %s mode...", serverName, serverVersion, *mode)

	// Create echo tool
	echoTool := &domain.Tool{
		Name:        "mcp_golang_mcp_server_ws_mcp_golang_mcp_server_sse_echo",
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

	// Use the builder pattern to create a base server configuration
	ctx := context.Background()
	serverBuilder := builder.NewServerBuilder().
		WithName(serverName).
		WithVersion(serverVersion).
		WithInstructions(serverInstructions).
		WithAddress(*addr).
		AddTool(ctx, echoTool)

	// Start the appropriate server based on mode
	switch *mode {
	case "http":
		// Build and start the HTTP server
		mcpServer := serverBuilder.BuildMCPServer()
		startHTTPServer(mcpServer, logger)

	case "stdio":
		// Start the StdIO server
		logger.Println("Starting StdIO server. Send JSON-RPC requests via stdin.")
		err := serverBuilder.ServeStdio(
			stdio.WithErrorLogger(logger),
		)
		if err != nil {
			logger.Fatalf("Error serving stdio: %v", err)
		}

	case "both":
		// Start both HTTP and StdIO servers
		mcpServer := serverBuilder.BuildMCPServer()

		// Start HTTP server in a goroutine
		go startHTTPServer(mcpServer, logger)

		// Start StdIO server in the main thread
		logger.Println("Starting StdIO server. Send JSON-RPC requests via stdin.")
		err := serverBuilder.ServeStdio(
			stdio.WithErrorLogger(logger),
		)
		if err != nil {
			logger.Fatalf("Error serving stdio: %v", err)
		}

	default:
		logger.Fatalf("Unknown mode: %s. Valid modes are http, stdio, or both", *mode)
	}
}

// startHTTPServer starts the HTTP server and handles graceful shutdown
func startHTTPServer(mcpServer *rest.MCPServer, logger *log.Logger) {
	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Printf("HTTP server starting on %s", mcpServer.GetAddress())
		if err := mcpServer.Start(); err != nil {
			if err.Error() != "http: Server closed" {
				logger.Fatalf("Server failed to start: %v", err)
			}
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logger.Println("Shutting down HTTP server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := mcpServer.Stop(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("HTTP server stopped gracefully")
}
