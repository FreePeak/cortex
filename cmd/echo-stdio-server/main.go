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
