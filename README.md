<h1 align="center">

   <img alt="main logo" src="logo.svg" width="150"/>
   
   <br/>
   Cortex
</h1>
<h4 align="center">Build MCP Servers Declaratively in Golang</h4>

<p align="center">
	<a href="https://pkg.go.dev/github.com/FreePeak/cortex"><img src="https://pkg.go.dev/badge/github.com/FreePeak/cortex.svg" alt="Go Reference"></a>
	<a href="https://goreportcard.com/report/github.com/FreePeak/cortex"><img src="https://goreportcard.com/badge/github.com/FreePeak/cortex" alt="Go Reference"></a>
	<a href="https://github.com/FreePeak/cortex/actions/workflows/test.yaml"><img src="https://github.com/FreePeak/cortex/actions/workflows/test.yaml/badge.svg"/></a>
	<a href="https://github.com/FreePeak/cortex/actions/workflows/golangci-lint.yaml"><img src="https://github.com/FreePeak/cortex/actions/workflows/golangci-lint.yaml/badge.svg"/></a>
</p>

#### Build MCP Servers Declaratively in Go

[Features](#features) • [Tool Example](#tool-example) • [Documentation](#documentation) • [Sources](#sources)

Cortex is a Go library for building context servers supporting the Model Context Protocol (MCP). It provides a clean, declarative approach to building MCP servers, allowing you to easily define tools, resources, and prompts.

With Cortex, you can easily integrate AI capabilities into your applications by providing a standardized way for LLMs to interact with your application's data and functionality.

## Features

Here is a list of features that are implemented and planned:

* **Base**
  * Lifecycle management
  * Health/ping endpoint
  * Progress tracking (planned)
* **Transports**
  * Stdio Transport
  * HTTP/SSE Transport
  * Multi-protocol support
* **Tools**
  * Declarative tool definition
  * Parameter validation
  * Type safety
* **Resources**
  * Static resources
  * Dynamic resources
  * Resource templates (planned)
  * Resource subscriptions
* **Prompts**
  * Prompt definitions
  * Prompt completion
* **Testing**
  * Functional testing utilities
  * Mocks for testing
* **Core Infrastructure**
  * Clean dependency injection
  * Structured logging
  * Graceful shutdown

## Tool Example

Here's a simple example of creating an echo server with Cortex:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
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
```

You can run this example with:

```bash
# Run the example
go run examples/echo_server.go
```

Test with:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","parameters":{"message":"Hello, World!"}}}' | go run examples/echo_server.go
```

## Calculator Example

For a more advanced example, check out the calculator server which demonstrates:
- Running in either HTTP or stdio mode
- Using command-line flags for configuration
- Parameter validation and type conversion
- Graceful shutdown for HTTP servers

```bash
# Run in HTTP mode
go run examples/calculator/main.go --mode http

# Run in stdio mode
go run examples/calculator/main.go --mode stdio
```

## Documentation

Comprehensive documentation is available in the [docs](./docs) directory.

## About MCP

The [Model Context Protocol](https://modelcontextprotocol.io) (MCP) is a standard protocol for providing context to Large Language Models. It allows LLMs to access data and functionality from applications in a secure, controlled way.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📧 Support & Contact

- For questions or issues, email [mnhatlinh.doan@gmail.com](mailto:mnhatlinh.doan@gmail.com)
- Open an issue directly: [Issue Tracker](https://github.com/FreePeak/cortex/issues)
- If Cortex helps your work, please consider supporting:

<p align="">
<a href="https://www.buymeacoffee.com/linhdmn">
<img src="https://img.buymeacoffee.com/button-api/?text=Support Cortex&emoji=☕&slug=linhdmn&button_colour=FFDD00&font_colour=000000&font_family=Cookie&outline_colour=000000&coffee_colour=ffffff" 
alt="Buy Me A Coffee"/>
</a>
</p>

