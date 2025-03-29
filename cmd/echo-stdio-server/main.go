package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys
const (
	timestampKey contextKey = "timestamp"
)

// Create a custom context function that adds a timestamp
func withTimestamp(ctx context.Context) context.Context {
	// Log when the context function is called - we'll use the logger
	// that will be attached to the context later
	return context.WithValue(ctx, timestampKey, fmt.Sprintf("%d", time.Now().Unix()))
}

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
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Echo Stdio Server...")

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
	logger.Info("Server ready. You can now send JSON-RPC requests via stdin.")
	err = serverBuilder.ServeStdio(
		stdio.WithLogger(logger),
		stdio.WithStdioContextFunc(withTimestamp),
	)

	if err != nil {
		logger.Error("Error serving stdio", logging.Fields{"error": err})
		os.Exit(1)
	}
}
