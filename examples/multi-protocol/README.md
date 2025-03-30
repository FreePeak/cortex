# Multi-Protocol MCP Server Example

This example demonstrates an MCP server that can communicate over either stdio or HTTP, depending on the configuration. It integrates with external tool providers to showcase the plugin architecture.

## Features

- Supports both stdio and HTTP communication protocols
- Integrates with external tool providers (weather and database)
- Command-line configurable settings
- Graceful shutdown support

## Running the Example

### Stdio Mode

```bash
go run examples/multi-protocol/main.go -protocol stdio
```

### HTTP Mode

```bash
go run examples/multi-protocol/main.go -protocol http -address localhost:8080
```

## Command-Line Options

- `-protocol`: Communication protocol (stdio or http, default: stdio)
- `-address`: HTTP server address when using http protocol (default: localhost:8080)

## Available Tools

This example integrates with two tool providers:

### Weather Provider

- `weather`: Gets today's weather forecast for a location
- `forecast`: Gets a multi-day weather forecast for a location

Example JSON-RPC request (stdio):

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "weather",
    "parameters": {
      "location": "New York"
    }
  }
}
```

### Database Provider

- `db.get`: Gets a value from the database by key
- `db.set`: Sets a value in the database
- `db.delete`: Deletes a value from the database
- `db.keys`: Lists all keys in the database

Example JSON-RPC request (stdio):

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "db.set",
    "parameters": {
      "key": "user1",
      "value": {
        "name": "John",
        "age": 30
      }
    }
  }
}
```

## Code Structure

- `main.go`: Sets up the server and registers the tool providers
- `examples/providers/weather`: Weather provider implementation
- `examples/providers/database`: Database provider implementation

## Implementation Details

The server is implemented using the Cortex MCP server platform. It uses the `server.NewMCPServer` function to create a server instance and registers tool providers using the `RegisterProvider` method. The server is configured to use either stdio or HTTP based on the command-line options. 