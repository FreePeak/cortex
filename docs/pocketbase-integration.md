# Integrating Cortex with PocketBase

This guide explains how to integrate Cortex into a [PocketBase](https://github.com/pocketbase/pocketbase) application to provide MCP (Model Context Protocol) capabilities.

## Overview

PocketBase is an open-source Go backend that provides a database, auth system, realtime subscriptions, and admin UI. By embedding Cortex into PocketBase, you can add MCP capabilities to your PocketBase application, allowing it to:

1. Expose tools for LLMs to call
2. Provide resources for LLMs to access
3. Support Model Context Protocol (MCP) communication

## Integration Methods

There are two primary ways to integrate Cortex with PocketBase:

1. **Plugin Integration**: Use Cortex's PocketBase plugin for seamless integration
2. **Manual Integration**: Manually add Cortex handlers to your PocketBase router

This guide focuses on the plugin integration approach, which is simpler and recommended for most users.

## Plugin Integration

### Prerequisites

- Go 1.20 or later
- A PocketBase application

### Step 1: Add Dependencies

Add Cortex to your PocketBase application's `go.mod` file:

```bash
go get github.com/FreePeak/cortex
```

### Step 2: Create a Cortex Plugin

In your PocketBase application, create a new file for the Cortex plugin integration:

```go
// cortex_plugin.go
package main

import (
    "log"
    "os"

    "github.com/FreePeak/cortex/pkg/integration/pocketbase"
    "github.com/FreePeak/cortex/pkg/tools"
)

func createCortexPlugin(app *pocketbase.PocketBase) {
    // Create a logger for the plugin
    logger := log.New(os.Stderr, "[cortex] ", log.LstdFlags)

    // Create the plugin with desired options
    plugin := pocketbase.NewCortexPlugin(
        pocketbase.WithLogger(logger),
        pocketbase.WithBasePath("/api/mcp"),
        pocketbase.WithName("My PocketBase MCP Server"),
        pocketbase.WithVersion("1.0.0"),
    )

    // Register tools
    echoTool := tools.NewTool("echo",
        tools.WithDescription("Echoes back the input message"),
        tools.WithString("message",
            tools.Description("The message to echo back"),
            tools.Required(),
        ),
    )

    // Register the tool with a handler
    plugin.AddTool(echoTool, func(ctx context.Context, request pocketbase.ToolCallRequest) (interface{}, error) {
        message := request.Parameters["message"].(string)
        return map[string]interface{}{
            "content": []map[string]interface{}{
                {
                    "type": "text",
                    "text": message,
                },
            },
        }, nil
    })

    // Register the plugin with PocketBase
    app.RegisterPlugin(plugin)
}
```

### Step 3: Use the Plugin in Your PocketBase App

Update your main application to use the Cortex plugin:

```go
// main.go
package main

import (
    "log"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

func main() {
    app := pocketbase.New()

    // Register hooks and other PocketBase configuration...

    // Register the Cortex plugin
    createCortexPlugin(app)

    // Start the PocketBase app
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Step 4: Access the MCP Server

Once your application is running, the MCP server will be available at the configured base path (default: `/api/mcp`). You can interact with it using MCP clients like the official [MCP JS SDK](https://github.com/Model-Context-Protocol/mcp-js).

## Authentication and Authorization

By default, the Cortex plugin uses PocketBase's authentication system. You can configure authentication requirements in the plugin options:

```go
plugin := pocketbase.NewCortexPlugin(
    // ...other options...
    pocketbase.WithAuth(pocketbase.AuthOptions{
        Required: true,                   // Require authentication for all MCP endpoints
        AdminOnly: false,                 // Allow regular users (not just admins)
        AllowedRoles: []string{"api"},    // Only allow users with the "api" role
    }),
)
```

## Adding Resources from PocketBase Collections

You can expose PocketBase collections as MCP resources:

```go
plugin.AddCollectionResource("tasks", "Sample Tasks", "List of example tasks")
```

This will make the collection available as an MCP resource that LLMs can access.

## Advanced Configuration

For more advanced configuration options, see the full API documentation in the Cortex repository.

## Example Application

A complete example application that demonstrates Cortex integration with PocketBase is available in the `examples/integration/pocketbase` directory of the Cortex repository.

## Troubleshooting

If you encounter issues with the integration:

1. Check the logs for error messages
2. Verify your PocketBase application is running correctly
3. Ensure the Cortex plugin is registered properly
4. Confirm that your tools are registered with the correct handlers

For more help, see the troubleshooting section in the main Cortex documentation. 