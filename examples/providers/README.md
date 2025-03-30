# Cortex Example Providers

This directory contains example implementations of Cortex tool providers that demonstrate how to integrate with the Cortex MCP server platform. These examples are designed to showcase the flexibility of the plugin architecture.

## Weather Provider

The Weather Provider (`weather` directory) demonstrates integrating a simple weather service:

- **weather tool**: Gets today's weather forecast for a location
- **forecast tool**: Gets a multi-day weather forecast for a location

Example usage (via JSON-RPC):

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

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "forecast",
    "parameters": {
      "location": "San Francisco",
      "days": 5
    }
  }
}
```

## Database Provider

The Database Provider (`database` directory) demonstrates implementing a simple in-memory key-value store:

- **db.get tool**: Gets a value from the database by key
- **db.set tool**: Sets a value in the database
- **db.delete tool**: Deletes a value from the database
- **db.keys tool**: Lists all keys in the database

Example usage (via JSON-RPC):

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "db.set",
    "parameters": {
      "key": "user1",
      "value": {
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
      }
    }
  }
}
```

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "db.get",
    "parameters": {
      "key": "user1"
    }
  }
}
```

## Running Example Providers

The example providers are integrated into the flexible server example. To run the server with these providers:

```bash
# Run with stdio protocol
./run-flexible-stdio.sh

# Run with HTTP protocol
./run-flexible-http.sh
```

## Creating Your Own Provider

To create your own provider, you can use these examples as templates. Here are the basic steps:

1. Define a struct that embeds `plugin.BaseProvider`
2. Create a constructor that initializes the provider and registers tools
3. Implement tool handler functions
4. Register your provider with the flexible MCP server

For a complete guide, see the README in the `pkg/plugin` directory.

## Best Practices

When creating your own providers, follow these best practices:

1. Use descriptive names for your provider and tools
2. Define clear parameter names and descriptions
3. Validate all input parameters
4. Return informative error messages
5. Follow the MCP protocol response format
6. Use context for cancellation and timeouts
7. Implement proper logging

## Extending These Examples

These examples are intentionally simple to demonstrate the core concepts. In a real-world implementation, you might want to:

- Connect to actual external services
- Implement caching for better performance
- Add authentication and authorization
- Implement rate limiting
- Add more sophisticated error handling
- Include monitoring and metrics 