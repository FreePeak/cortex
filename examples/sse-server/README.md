# SSE MCP Server Example

This example demonstrates an MCP server that communicates over HTTP using Server-Sent Events (SSE). It provides two basic tools: an echo tool and a weather forecast tool.

## Features

- Communicates using HTTP and Server-Sent Events
- Provides an echo tool that reflects messages back to the client
- Provides a weather tool that generates random weather forecasts
- Supports graceful shutdown

## Running the Example

```bash
go run examples/sse-server/main.go
```

By default, the server listens on port 8080.

## Usage

Once the server is running, you can connect to it using any HTTP client that supports SSE. For example, you can use curl:

```bash
curl -X POST http://localhost:8080/tools/call -H "Content-Type: application/json" -d '{"name":"echo","parameters":{"message":"Hello, World!"}}'
```

### Integration with Cursor

You can connect to this server from Cursor by going to Settings > Extensions > Model Context Protocol and entering `http://localhost:8080` as the server URL.

### Available Tools

#### Echo Tool

Example HTTP request:

```bash
curl -X POST http://localhost:8080/tools/call -H "Content-Type: application/json" -d '{"name":"echo","parameters":{"message":"Hello, World!"}}'
```

#### Weather Tool

Example HTTP request:

```bash
curl -X POST http://localhost:8080/tools/call -H "Content-Type: application/json" -d '{"name":"weather","parameters":{"location":"New York"}}'
```

## Code Structure

- `main.go`: Sets up the server and registers the tools
- Tool handlers:
  - `handleEcho`: Processes echo tool requests
  - `handleWeather`: Processes weather tool requests

## Implementation Details

The server is implemented using the Cortex MCP server platform. It uses the `server.NewMCPServer` function to create a server instance and registers tools using the `AddTool` method. The server is configured to listen on HTTP and uses graceful shutdown to ensure all connections are properly closed when the server is terminated. 