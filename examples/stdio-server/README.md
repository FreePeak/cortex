# Stdio MCP Server Example

This example demonstrates a simple MCP server that communicates over standard input/output. It provides two basic tools: an echo tool and a weather forecast tool.

## Features

- Communicates using the JSON-RPC protocol over standard I/O
- Provides an echo tool that reflects messages back to the client
- Provides a weather tool that generates random weather forecasts

## Running the Example

```bash
go run examples/stdio-server/main.go
```

## Usage

Once the server is running, you can send JSON-RPC requests via standard input. Here are some example requests:

### Echo Tool

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","parameters":{"message":"Hello, World!"}}}
```

This will echo back the message "Hello, World!" with a timestamp.

### Weather Tool

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"weather","parameters":{"location":"New York"}}}
```

This will return a random weather forecast for New York.

## Code Structure

- `main.go`: Sets up the server and registers the tools
- Tool handlers:
  - `handleEcho`: Processes echo tool requests
  - `handleWeather`: Processes weather tool requests

## Implementation Details

The server is implemented using the Cortex MCP server platform. It uses the `server.NewMCPServer` function to create a server instance and registers tools using the `AddTool` method. Tool handlers are implemented as functions that take a context and a `ToolCallRequest` and return a response. 