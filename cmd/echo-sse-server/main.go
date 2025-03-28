package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

const (
	serverName       = "Example MCP Server"
	serverVersion    = "0.1.0"
	serverAddr       = ":8080"
	shutdownTimeout  = 10 * time.Second
	shutdownGraceful = 2 * time.Second

	serverInstructions = `
This is an example MCP server with sample resources, tools, and prompts.

Available resources:
- sample://hello-world: A hello world resource

Available tools:
- echo: Echoes back the input message

Available prompts:
- greeting: A simple greeting prompt

You can connect to this server from Cursor by going to Settings > Extensions >
Model Context Protocol and entering 'http://localhost:8080' as the server URL.
`
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Example MCP Server...")

	// Create sample data
	sampleResource := &domain.Resource{
		URI:         "sample://hello-world",
		Name:        "Hello World Resource",
		Description: "A sample resource for demonstration purposes",
		MIMEType:    "text/plain",
	}

	sampleTool := &domain.Tool{
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

	samplePrompt := &domain.Prompt{
		Name:        "greeting",
		Description: "A simple greeting prompt",
		Template:    "Hello, {{name}}! Welcome to {{place}}.",
		Parameters: []domain.PromptParameter{
			{
				Name:        "name",
				Description: "The name to greet",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "place",
				Description: "The place to welcome to",
				Type:        "string",
				Required:    true,
			},
		},
	}

	// Use the builder pattern to create the server
	ctx := context.Background()
	serverBuilder := builder.NewServerBuilder().
		WithName(serverName).
		WithVersion(serverVersion).
		WithInstructions(serverInstructions).
		WithAddress(serverAddr).
		AddResource(ctx, sampleResource).
		AddTool(ctx, sampleTool).
		AddPrompt(ctx, samplePrompt)

	// Build the MCP server
	mcpServer := serverBuilder.BuildMCPServer()

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := mcpServer.Start(); err != nil {
			if err.Error() != "http: Server closed" {
				log.Fatalf("Server failed to start: %v", err)
			}
		}
	}()

	log.Printf("Server is running on %s", serverAddr)
	log.Printf("You can connect to this server from Cursor by going to Settings > Extensions > Model Context Protocol and entering 'http://localhost:8080' as the server URL.")
	log.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := mcpServer.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Small delay to allow final cleanup
	time.Sleep(shutdownGraceful)
	log.Println("Server stopped gracefully")
}
