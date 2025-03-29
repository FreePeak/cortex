// This example demonstrates how to use the MCP Server SDK to create
// a server that can run in either HTTP or stdio mode with a calculator tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/pkg/server"
	"github.com/FreePeak/golang-mcp-server-sdk/pkg/tools"
)

const (
	serverName      = "Calculator MCP Server"
	serverVersion   = "1.0.0"
	shutdownTimeout = 10 * time.Second
)

func main() {
	// Parse command-line flags
	mode := flag.String("mode", "http", "Server mode: http or stdio")
	addr := flag.String("addr", ":8080", "HTTP server address (for HTTP mode)")
	flag.Parse()

	// Create the MCP server
	mcpServer := server.NewMCPServer(serverName, serverVersion)
	mcpServer.SetAddress(*addr)

	// Create a calculator tool with parameters
	calculatorTool := tools.NewTool("calculator",
		tools.WithDescription("Performs basic arithmetic operations"),
		tools.WithString("operation",
			tools.Description("The operation to perform (add, subtract, multiply, divide)"),
			tools.Required(),
		),
		tools.WithNumber("a",
			tools.Description("First operand"),
			tools.Required(),
		),
		tools.WithNumber("b",
			tools.Description("Second operand"),
			tools.Required(),
		),
	)

	// Add the tool with its handler
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, calculatorTool, handleCalculator)
	if err != nil {
		log.Fatalf("Failed to add calculator tool: %v", err)
	}

	log.Printf("Starting %s v%s in %s mode", serverName, serverVersion, *mode)

	// Start the server in the specified mode
	switch *mode {
	case "http":
		// Start HTTP server with graceful shutdown
		startHTTPServer(mcpServer)

	case "stdio":
		// Start stdio server
		log.Println("Stdio server started. Send JSON-RPC requests via stdin.")
		log.Println("Example: {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"calculator\",\"parameters\":{\"operation\":\"add\",\"a\":5,\"b\":3}}}")
		if err := mcpServer.ServeStdio(); err != nil {
			log.Fatalf("Stdio server error: %v", err)
		}

	default:
		log.Fatalf("Unknown mode: %s. Valid modes are 'http' or 'stdio'", *mode)
	}
}

// startHTTPServer starts the HTTP server with graceful shutdown
func startHTTPServer(mcpServer *server.MCPServer) {
	// Create a shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("HTTP server starting at http://localhost%s", mcpServer.GetAddress())
		log.Println("You can connect to this server from Cursor by configuring the Model Context Protocol extension.")
		if err := mcpServer.ServeHTTP(); err != nil {
			if err.Error() != "http: Server closed" {
				log.Fatalf("HTTP server error: %v", err)
			}
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutting down HTTP server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown the server
	if err := mcpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown error: %v", err)
	}

	log.Println("HTTP server stopped gracefully")
}

// handleCalculator handles calculator tool calls
func handleCalculator(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract parameters
	operation, ok := request.Parameters["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid operation parameter")
	}

	// Extract and convert number parameters
	aParam, aOk := request.Parameters["a"]
	bParam, bOk := request.Parameters["b"]
	if !aOk || !bOk {
		return nil, fmt.Errorf("missing operands")
	}

	// Convert parameters to float64
	var a, b float64
	var err error

	switch v := aParam.(type) {
	case float64:
		a = v
	case int:
		a = float64(v)
	case string:
		a, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid first operand: %v", err)
		}
	default:
		return nil, fmt.Errorf("first operand has unsupported type")
	}

	switch v := bParam.(type) {
	case float64:
		b = v
	case int:
		b = float64(v)
	case string:
		b, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid second operand: %v", err)
		}
	default:
		return nil, fmt.Errorf("second operand has unsupported type")
	}

	// Perform the calculation
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	// Format the result
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Result of %s %v %v = %v", operation, a, b, result),
			},
		},
	}, nil
}
