# PocketBase Integration Example

This example demonstrates how to integrate Cortex with [PocketBase](https://github.com/pocketbase/pocketbase), allowing you to add MCP capabilities to your PocketBase applications.

## Overview

In this example, we're using a mock PocketBase server since we don't want to add PocketBase as a dependency to the Cortex repository. In a real application, you would replace the mock with the actual PocketBase library.

The example shows:

1. How to create a Cortex plugin for PocketBase
2. How to configure the plugin with custom options
3. How to add tools to the plugin
4. How to register the plugin with PocketBase

## Running the Example

To run this example:

```bash
go run examples/integration/pocketbase/main.go
```

This will start a mock PocketBase server with the Cortex plugin registered at the `/api/mcp` path.

By default, the server runs on port 8080. If this port is already in use, you can specify a different port:

```bash
go run examples/integration/pocketbase/main.go --port 8090
```

You can also specify a custom data directory for PocketBase:

```bash
go run examples/integration/pocketbase/main.go --data ./my_data_dir
```

### Using the Run Script

For convenience, you can also use the included run script:

```bash
# Run with default settings
./run.sh

# Run with custom port
./run.sh --port 8090

# Run with custom data directory
./run.sh --data ./my_data_dir

# Kill any existing processes on the port before starting
./run.sh --kill

# Show help
./run.sh --help
```

### Troubleshooting Connection Issues

If you encounter connection issues:

1. **Port in use**: If port 8080 is already in use, use the `--port` flag to specify a different port
2. **404 errors with SSE connections**: Ensure your client is using the correct URL path (`/api/mcp/sse`)
3. **CORS issues**: The server is configured with permissive CORS headers, but some browsers might still block requests. Consider using a CORS extension or proxy if needed

To kill any processes using port 8080, you can use the included script:

```bash
# Make the script executable
chmod +x kill-port.sh

# Kill processes on the default port (8080)
./kill-port.sh

# Or specify a different port
./kill-port.sh 8090
```

### Connecting with Cursor Editor

To connect the Cursor editor to this example:

1. Run the server with: `go run examples/integration/pocketbase/main.go`
2. In Cursor, open the MCP settings
3. Set the MCP URL to: `http://localhost:8080/api/mcp` (or with your custom port)
4. Test the connection using the Echo tool

### Testing the Connection

You can test the server using the included test clients:

#### Browser Test (HTML UI)

1. Start the server with: `go run examples/integration/pocketbase/main.go`
2. Open the `test-client.html` file in your browser
3. Use the interactive UI to test different aspects of the connection

#### JavaScript Console Test

1. Start the server with: `go run examples/integration/pocketbase/main.go`
2. Open `test-client.js` in a browser console or import it in your HTML
3. Check the console for test results

#### Node.js Test

1. Start the server with: `go run examples/integration/pocketbase/main.go`
2. Run the test client: `node test-client.js`

All test clients will verify:
- Basic connectivity to the server
- Echo tool functionality
- Weather tool functionality
- SSE connection capability

If you're using a custom port, make sure to update the port settings in the test clients.

## Key Components

### CortexPlugin

The `CortexPlugin` is the main integration point between Cortex and PocketBase. It:

- Wraps an `MCPServer` to provide MCP functionality
- Exposes methods for adding tools and registering with PocketBase
- Provides HTTP handlers that can be registered with PocketBase

### Tool Handlers

The example includes two tool handlers:

1. **Echo Tool**: Echoes back the input message
2. **Weather Tool**: Simulates getting a weather forecast for a location

### PocketBase Integration

In a real PocketBase application, you would integrate Cortex like this:

```go
import (
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/FreePeak/cortex/pkg/integration/pocketbase"
)

func main() {
    app := pocketbase.New()
    
    // Initialize Cortex plugin
    plugin := pocketbase.NewCortexPlugin(
        pocketbase.WithName("PocketBase MCP Server"),
        pocketbase.WithVersion("1.0.0"),
        pocketbase.WithBasePath("/api/mcp"),
    )
    
    // Add tools to the plugin
    // ...
    
    // Register the plugin with PocketBase
    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        // Get the Cortex HTTP handler
        handler := plugin.GetHTTPHandler()
        
        // Register it with PocketBase's router
        e.Router.ANY("/api/mcp/*", handler)
        
        return nil
    })
    
    // Start the PocketBase app
    if err := app.Start(); err != nil {
        panic(err)
    }
}
```

## Learn More

For more detailed documentation on integrating Cortex with PocketBase, see the [PocketBase Integration Guide](../../../docs/pocketbase-integration.md).