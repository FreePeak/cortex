package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

// Record a timestamp for demo purposes
func getTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func main() {
	// Create the server with name and version
	mcpServer := server.NewMCPServer("Echo Stdio Server", "1.0.0")

	// Create the echo tool using the fluent API
	echoTool := tools.NewTool("echo_golang_mcp_server_stdio",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Add the tool with a handler function
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		log.Fatalf("Error adding tool: %v", err)
	}

	// Print server ready message
	fmt.Println("Server ready. You can now send JSON-RPC requests via stdin.")
	fmt.Println("Try: {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"echo_golang_mcp_server_stdio\",\"parameters\":{\"message\":\"Hello, World!\"}}}")

	// Start the server
	if err := mcpServer.ServeStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "Error serving stdio: %v\n", err)
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

	// Add a timestamp to show we can process the message
	timestamp := getTimestamp()
	responseMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// Return the echo response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": responseMessage,
			},
		},
	}, nil
}
