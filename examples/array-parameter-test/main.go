package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
	"github.com/FreePeak/cortex/pkg/types"
)

func main() {
	// Create a logger that writes to stderr
	logger := log.New(os.Stderr, "[array-test] ", log.LstdFlags)

	// Create the server
	mcpServer := server.NewMCPServer("Array Parameter Test", "1.0.0", logger)

	// Create a tool with array parameter
	arrayTool := tools.NewTool("array_test",
		tools.WithDescription("Test tool with array parameter"),
		tools.WithArray("string_array",
			tools.Description("Array of strings"),
			tools.Required(),
			tools.Items(map[string]interface{}{
				"type": "string",
			}),
		),
		tools.WithArray("number_array",
			tools.Description("Array of numbers"),
			tools.Items(map[string]interface{}{
				"type": "number",
			}),
		),
	)

	// Add the tool to the server
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, arrayTool, handleArrayTest)
	if err != nil {
		logger.Fatalf("Error adding tool: %v", err)
	}

	// Print tool schema for debugging
	printToolSchema(arrayTool)

	// Write server status to stderr
	fmt.Fprintf(os.Stderr, "Starting Array Parameter Test Server...\n")
	fmt.Fprintf(os.Stderr, "Send JSON-RPC messages via stdin to interact with the server.\n")
	fmt.Fprintf(os.Stderr, `Try: {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"array_test","parameters":{"string_array":["a","b","c"]}}}\n`)

	// Serve over stdio
	if err := mcpServer.ServeStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Handler for the array test tool
func handleArrayTest(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract the string array parameter
	stringArray, ok := request.Parameters["string_array"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'string_array' parameter")
	}

	// Get the optional number array parameter
	var numberArray []interface{}
	if val, ok := request.Parameters["number_array"]; ok {
		numberArray, _ = val.([]interface{})
	}

	// Return the arrays in the response
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Received string array: %v\nReceived number array: %v", stringArray, numberArray),
			},
		},
	}, nil
}

// Print the tool schema
func printToolSchema(tool *types.Tool) {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	for _, param := range tool.Parameters {
		paramSchema := map[string]interface{}{
			"type":        param.Type,
			"description": param.Description,
		}

		if param.Type == "array" && param.Items != nil {
			paramSchema["items"] = param.Items
		}

		schema["properties"].(map[string]interface{})[param.Name] = paramSchema
	}

	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Fprintf(os.Stderr, "Tool schema:\n%s\n", schemaJSON)
}
