// Example of using the MCP SDK to create a simple echo server
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/pkg/server"
	"github.com/FreePeak/golang-mcp-server-sdk/pkg/tools"
)

func main() {
	// Create the server
	mcpServer := server.NewMCPServer("Echo Server Example", "1.0.0")

	// Create an echo tool
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Add the tool to the server with a handler
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		log.Fatalf("Error adding tool: %v", err)
	}

	// Start the server
	fmt.Println("Starting Echo Server...")
	fmt.Println("Send JSON-RPC messages via stdin to interact with the server.")
	fmt.Println("Try: {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"echo\",\"parameters\":{\"message\":\"Hello, World!\"}}}")

	// Serve over stdio
	if err := mcpServer.ServeStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
