package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
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

	// Use the new server builder pattern
	ctx := context.Background()
	serverBuilder := builder.NewServerBuilder().
		WithName("Echo Stdio Server").
		WithVersion("1.0.0").
		WithInstructions("This is a simple echo server that echoes back messages sent to it.").
		WithAddress(":0"). // Dummy address since we won't be using HTTP
		AddTool(ctx, echoTool)

	// Start the stdio server with our custom context function
	logger.Println("Server ready. You can now send JSON-RPC requests via stdin.")
	err := serverBuilder.ServeStdio(
		stdio.WithErrorLogger(logger),
		stdio.WithStdioContextFunc(withTimestamp),
	)

	if err != nil {
		logger.Printf("Error serving stdio: %v", err)
		os.Exit(1)
	}
}
