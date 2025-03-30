# Cortex Plugin System

The Cortex Plugin System provides a flexible architecture for extending the Cortex MCP (Model Context Protocol) server with external tools and services. This system allows third-party providers to register tools that can be called through the MCP protocol.

## Overview

The plugin system includes the following key components:

- **Provider Interface**: Defines the contract that tool providers must implement
- **Registry**: Manages the registration and discovery of providers and tools
- **Base Provider**: Provides a foundation for building tool providers

## Provider Interface

Tool providers must implement the `Provider` interface:

```go
type Provider interface {
    // GetProviderInfo returns information about the tool provider
    GetProviderInfo(ctx context.Context) (*ProviderInfo, error)

    // GetTools returns a list of tools provided by this provider
    GetTools(ctx context.Context) ([]*types.Tool, error)

    // ExecuteTool executes a specific tool with the given parameters
    ExecuteTool(ctx context.Context, request *ExecuteRequest) (*ExecuteResponse, error)
}
```

## Creating a Tool Provider

The easiest way to create a tool provider is to use the `BaseProvider` implementation:

```go
// Create provider info
info := plugin.ProviderInfo{
    ID:          "my-provider",
    Name:        "My Provider",
    Version:     "1.0.0",
    Description: "A custom tool provider",
    Author:      "Your Name",
    URL:         "https://github.com/yourusername/your-repo",
}

// Create base provider
baseProvider := plugin.NewBaseProvider(info, logger)

// Create your custom provider
myProvider := &MyProvider{
    BaseProvider: baseProvider,
    // Add your custom fields here
}

// Register tools with your provider
myTool := tools.NewTool("my-tool",
    tools.WithDescription("A custom tool"),
    tools.WithString("param1", tools.Description("Parameter 1"), tools.Required()),
)

// Register the tool with your provider
myProvider.RegisterTool(myTool, handleMyTool)
```

## Tool Handler Function

Each tool needs a handler function that follows this signature:

```go
func handleMyTool(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
    // Extract parameters
    param1, ok := params["param1"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'param1' parameter")
    }

    // Process the request
    result := fmt.Sprintf("Processed param1: %s", param1)

    // Return the result in the format expected by the MCP protocol
    return map[string]interface{}{
        "content": []map[string]interface{}{
            {
                "type": "text",
                "text": result,
            },
        },
    }, nil
}
```

## Using the Registry

The registry manages the providers and tools:

```go
// Create a registry
registry := plugin.NewRegistry(logger)

// Register a provider
registry.RegisterProvider(ctx, myProvider)

// Get a tool
tool, provider, err := registry.GetTool(ctx, "my-tool")
if err != nil {
    // Handle error
}

// List all tools
tools, err := registry.ListTools(ctx)
if err != nil {
    // Handle error
}
```

## Using the Flexible MCP Server

The `FlexibleMCPServer` supports dynamic tool providers:

```go
// Create a registry
registry := plugin.NewRegistry(logger)

// Create the flexible MCP server
mcpServer := server.NewFlexibleMCPServer("My MCP Server", "1.0.0", registry, logger)

// Register providers with the server
mcpServer.RegisterProvider(ctx, myProvider)

// Start the server (stdio or HTTP)
mcpServer.ServeStdio()
// or
mcpServer.ServeHTTP()
```

## Example Providers

The Cortex project includes example providers in the `examples/providers/` directory:

- **Weather Provider**: Provides tools for getting weather forecasts
- **Database Provider**: Provides tools for simple database operations

These examples demonstrate how to create providers that integrate with the Cortex platform.

## Security Considerations

When implementing tool providers, consider the following security best practices:

1. Validate all input parameters thoroughly
2. Limit access to sensitive operations
3. Use context for cancellation and timeouts
4. Log security-relevant events
5. Avoid exposing internal details in error messages 