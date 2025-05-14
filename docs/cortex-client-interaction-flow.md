# Cortex MCP Server: Client Interaction Flow

This document describes how the Cortex MCP server interacts with client applications such as Cursor IDE and Claude Desktop.

## Architecture Overview

```mermaid
flowchart TD
    Client["Client Application\n(Cursor IDE / Claude Desktop)"]
    MCPServer["Cortex MCP Server"]
    ToolReg["Tool Registry\n(plugin.Registry)"]
    ToolHandlers["Tool Handlers\n(ToolHandler)"]
    NotifSystem["Notification System"]
    ToolProviders["Tool Providers\n(plugin.Provider)"]
    ExtServices["External Services"]
    Database["Database\n(via Provider)"]
    
    Client -- "JSON-RPC over HTTP" --> MCPServer
    MCPServer -- "Server-Sent Events (SSE)" --> Client
    
    MCPServer <--> ToolReg
    MCPServer <--> ToolHandlers
    MCPServer <--> NotifSystem
    
    ToolReg <--> ToolProviders
    ToolProviders <--> ExtServices
    ToolProviders <--> Database
    
    subgraph "Core Server Components"
        MCPServer
        ToolReg
        ToolHandlers
        NotifSystem
    end
    
    subgraph "External Integrations"
        ToolProviders
        ExtServices
        Database
    end
```

## Connection Establishment

1. **Client Initiates Connection**:
   - The client connects to the Cortex MCP server via HTTP
   - Supported transport protocols:
     * HTTP API (REST)
     * Server-Sent Events (SSE) for real-time updates
     * Standard I/O (for command-line integration)

2. **Session Registration**:
   - Server assigns a unique session ID to the client
   - Client's user agent is recorded
   - A notification channel is established for server-to-client communication

## Communication Protocol (JSON-RPC)

### Client-to-Server Communication

Clients send JSON-RPC messages to the server's message endpoint:

```json
{
  "jsonrpc": "2.0",
  "id": "request-123",
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "parameters": { /* tool-specific parameters */ }
  }
}
```

### Server-to-Client Communication

Server responds with JSON-RPC responses:

```json
{
  "jsonrpc": "2.0",
  "id": "request-123",
  "result": { /* tool-specific response */ }
}
```

For errors:

```json
{
  "jsonrpc": "2.0",
  "id": "request-123",
  "error": {
    "code": -32000,
    "message": "Error message"
  }
}
```

### Real-time Updates via SSE

- Server sends notifications using Server-Sent Events (SSE)
- Notifications follow JSON-RPC format but without an ID field
- Used for status updates, progress reports, etc.

## Tool Execution Flow

```mermaid
sequenceDiagram
    participant Client
    participant Server as Cortex MCP Server
    participant Handler as Tool Handler
    participant Provider as Tool Provider
    
    Client->>Server: Connect & Register Session
    Server->>Client: Session ID & Endpoints
    
    Client->>Server: Tool Call Request
    Server->>Server: Validate Parameters
    
    alt Direct Tool
        Server->>Handler: Execute Tool
        Handler->>Server: Tool Result
    else Provider-based Tool
        Server->>Provider: Execute Tool
        Provider->>Provider: Process Request
        Provider->>Server: Tool Result
    end
    
    Server->>Client: JSON-RPC Response
    
    loop Real-time Updates
        Server->>Client: SSE Notifications
    end
    
    Client->>Server: Disconnect
    Server->>Server: Cleanup Session
```

## Integration Modes

1. **Standalone Server**:
   - Server runs independently and listens on a port
   - Provides HTTP API and SSE endpoints

2. **Embedded Mode**:
   - Server is embedded within another application (e.g., PocketBase)
   - Uses HTTP adapter for integration with existing HTTP servers

3. **STDIO Mode**:
   - Server communicates via standard input/output
   - Useful for CLI tools and scripted interactions

## Key Components

```mermaid
classDiagram
    class MCPServer {
        +AddTool()
        +RegisterProvider()
        +ServeHTTP()
        +ServeStdio()
        +ExecuteTool()
        +RegisterSession()
    }
    
    class SSEServer {
        +handleSSE()
        +handleMessage()
        +SendEventToSession()
        +BroadcastEvent()
    }
    
    class NotificationSender {
        +RegisterSession()
        +UnregisterSession()
        +SendNotification()
        +BroadcastNotification()
    }
    
    class ConnectionPool {
        +Add()
        +Remove()
        +Get()
        +Broadcast()
    }
    
    class ToolProvider {
        +GetTools()
        +ExecuteTool()
    }
    
    MCPServer --> SSEServer
    MCPServer --> NotificationSender
    SSEServer --> ConnectionPool
    MCPServer --> ToolProvider
    
    note for MCPServer "Core server implementation"
    note for SSEServer "Real-time communication"
    note for NotificationSender "Client notifications"
    note for ConnectionPool "Client session management"
    note for ToolProvider "Dynamic tool registration"
```

This architecture enables clients like Cursor IDE or Claude Desktop to interact with tools and services provided by the Cortex server, maintaining persistent connections for real-time updates. 