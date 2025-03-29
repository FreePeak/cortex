package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/pkg/server"
	"github.com/FreePeak/golang-mcp-server-sdk/pkg/tools"
)

const (
	serverName      = "Multi-Protocol MCP Server"
	serverVersion   = "1.0.0"
	defaultAddr     = ":8080"
	shutdownTimeout = 10 * time.Second

	serverInstructions = `
This is a multi-protocol MCP server example that can run in HTTP, SSE, or StdIO mode.
It demonstrates how to use the SDK to create different server types
with a shared configuration.

Available tool:
- echo_multi_protocol: Echoes back the input message
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

	// Create a server
	mcpServer := server.NewMCPServer(serverName, serverVersion)
	mcpServer.SetAddress(*addr)

	// Create echo tool using the fluent API
	echoTool := tools.NewTool("echo_multi_protocol",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Register the tool with its handler
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		logger.Fatalf("Failed to add tool: %v", err)
	}

	// Start the appropriate server based on mode
	switch *mode {
	case "http":
		// Start the HTTP server
		startHTTPServer(mcpServer, logger)

	case "stdio":
		// Start the StdIO server
		logger.Println("Starting StdIO server. Send JSON-RPC requests via stdin.")
		err := mcpServer.ServeStdio()
		if err != nil {
			logger.Fatalf("Error serving stdio: %v", err)
		}

	case "both":
		// Start both HTTP and StdIO servers
		// Start HTTP server in a goroutine
		go startHTTPServer(mcpServer, logger)

		// Start StdIO server in the main thread
		logger.Println("Starting StdIO server. Send JSON-RPC requests via stdin.")
		err := mcpServer.ServeStdio()
		if err != nil {
			logger.Fatalf("Error serving stdio: %v", err)
		}

	default:
		logger.Fatalf("Unknown mode: %s. Valid modes are http, stdio, or both", *mode)
	}
}

// handleEcho handles echo tool calls
func handleEcho(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract message parameter
	message, ok := request.Parameters["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid message parameter")
	}

	// Get current timestamp
	timestamp := time.Now().Format(time.RFC3339)

	// Return the echo response in MCP format
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Echo: %s (at %s)", message, timestamp),
			},
		},
	}, nil
}

// startHTTPServer starts the HTTP server and handles graceful shutdown
func startHTTPServer(mcpServer *server.MCPServer, logger *log.Logger) {
	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Printf("HTTP server starting on %s", mcpServer.GetAddress())
		if err := mcpServer.ServeHTTP(); err != nil {
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
	if err := mcpServer.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("HTTP server stopped gracefully")
}
