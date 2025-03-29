package main

import (
	"context"
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
	serverName       = "Example SSE MCP Server"
	serverVersion    = "1.0.0"
	serverAddr       = ":8080"
	shutdownTimeout  = 10 * time.Second
	shutdownGraceful = 2 * time.Second
)

func main() {
	// Create a new server using the SDK
	mcpServer := server.NewMCPServer(serverName, serverVersion)

	// Set the server address
	mcpServer.SetAddress(serverAddr)

	// Create tools with the fluent API
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Add tool with handler
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		log.Fatalf("Error adding tool: %v", err)
	}

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server is running on %s\n", serverAddr)
		fmt.Printf("You can connect to this server from Cursor by going to Settings > Extensions > Model Context Protocol and entering 'http://localhost%s' as the server URL.\n", serverAddr)
		fmt.Println("Press Ctrl+C to stop")

		// Use the SDK's built-in HTTP server functionality
		if err := mcpServer.ServeHTTP(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	fmt.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := mcpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Small delay to allow final cleanup
	time.Sleep(shutdownGraceful)
	fmt.Println("Server stopped gracefully")
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
