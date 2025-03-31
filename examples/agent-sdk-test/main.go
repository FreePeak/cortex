package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

func main() {
	// Create a logger that writes to stderr
	logger := log.New(os.Stderr, "[agent-sdk-test] ", log.LstdFlags)

	// Create the server
	mcpServer := server.NewMCPServer("Agent SDK Test", "1.0.0", logger)

	// Configure HTTP address
	mcpServer.SetAddress(":9095")

	// Create a tool with array parameter (compatible with OpenAI Agent SDK)
	queryTool := tools.NewTool("query_database",
		tools.WithDescription("Execute SQL query on a database"),
		tools.WithString("query",
			tools.Description("SQL query to execute"),
			tools.Required(),
		),
		tools.WithArray("params",
			tools.Description("Query parameters"),
			tools.Items(map[string]interface{}{
				"type": "string",
			}),
		),
	)

	// Add tool to the server
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, queryTool, handleQuery)
	if err != nil {
		logger.Fatalf("Error adding tool: %v", err)
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Printf("Starting Agent SDK Test server on %s", mcpServer.GetAddress())
		logger.Printf("Use the following URL in your OpenAI Agent SDK configuration: http://localhost:9095/sse")

		if err := mcpServer.ServeHTTP(); err != nil {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown gracefully
	logger.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mcpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server shutdown error: %v", err)
	}

	logger.Println("Server shutdown complete")
}

// Handler for the query tool
func handleQuery(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract the query parameter
	query, ok := request.Parameters["query"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'query' parameter")
	}

	// Get optional parameters
	var params []interface{}
	if paramsVal, ok := request.Parameters["params"]; ok {
		params, _ = paramsVal.([]interface{})
	}

	// In a real implementation, you would execute the query with the parameters
	// For this example, we'll just return mock data

	// Log the request
	log.Printf("Query received: %s", query)
	log.Printf("Parameters: %v", params)

	// Return a mock response
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Executed query: %s\nParameters: %v\n\nID\tName\tValue\n1\tItem1\t100\n2\tItem2\t200", query, params),
			},
		},
	}, nil
}
