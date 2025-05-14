// Package pocketbase provides integration between Cortex and PocketBase.
package pocketbase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/types"
)

// CortexPlugin is a PocketBase plugin that provides Cortex MCP server capabilities.
type CortexPlugin struct {
	name      string
	version   string
	basePath  string
	logger    *log.Logger
	mcpServer server.Embeddable
	port      int
}

// ToolCallRequest represents a request to execute a tool.
type ToolCallRequest struct {
	Name       string
	Parameters map[string]interface{}
	Session    *types.ClientSession
}

// Option is a function that configures a CortexPlugin.
type Option func(*CortexPlugin)

// WithName sets the name of the Cortex server.
func WithName(name string) Option {
	return func(p *CortexPlugin) {
		p.name = name
	}
}

// WithVersion sets the version of the Cortex server.
func WithVersion(version string) Option {
	return func(p *CortexPlugin) {
		p.version = version
	}
}

// WithBasePath sets the base path for the Cortex server routes.
func WithBasePath(basePath string) Option {
	return func(p *CortexPlugin) {
		p.basePath = basePath
	}
}

// WithLogger sets the logger for the Cortex server.
func WithLogger(logger *log.Logger) Option {
	return func(p *CortexPlugin) {
		p.logger = logger
	}
}

// WithPort sets the port for the Cortex server.
func WithPort(port int) Option {
	return func(p *CortexPlugin) {
		p.port = port
	}
}

// NewCortexPlugin creates a new CortexPlugin with the given options.
func NewCortexPlugin(opts ...Option) *CortexPlugin {
	plugin := &CortexPlugin{
		name:     "Cortex MCP Server",
		version:  "1.0.0",
		basePath: "/api/mcp",
		logger:   log.Default(),
		port:     8080,
	}

	// Apply options
	for _, opt := range opts {
		opt(plugin)
	}

	// Create the MCP server
	mcpServer := server.NewMCPServer(plugin.name, plugin.version, plugin.logger)

	// Set the server port address
	if plugin.port != 0 {
		address := fmt.Sprintf(":%d", plugin.port)
		plugin.logger.Printf("Setting MCP server address to: %s", address)
		mcpServer.SetAddress(address)
	}

	// Enable embed mode to prevent the server from binding to a port
	mcpServer.SetEmbedMode(true)

	// Store the server as an Embeddable interface
	plugin.mcpServer = mcpServer

	return plugin
}

// AddTool adds a tool to the Cortex server.
func (p *CortexPlugin) AddTool(tool *types.Tool, handler func(ctx context.Context, request ToolCallRequest) (interface{}, error)) error {
	// Add inputSchema to tool based on parameters
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, param := range tool.Parameters {
		// Add to properties
		paramSchema := map[string]interface{}{
			"type":        param.Type,
			"description": param.Description,
		}

		if param.Items != nil {
			paramSchema["items"] = param.Items
		}

		properties[param.Name] = paramSchema

		// Add to required if needed
		if param.Required {
			required = append(required, param.Name)
		}
	}

	// Set the inputSchema
	tool.InputSchema = map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}

	// Create an adapter to convert our tool handler to the MCP server's tool handler
	handlerAdapter := func(ctx context.Context, req server.ToolCallRequest) (interface{}, error) {
		// Convert the request
		request := ToolCallRequest{
			Name:       req.Name,
			Parameters: req.Parameters,
			Session:    req.Session,
		}

		// Call the handler
		return handler(ctx, request)
	}

	// Add the tool to the MCP server
	return p.mcpServer.AddTool(context.Background(), tool, handlerAdapter)
}

// RegisterWithPocketBase registers the Cortex plugin with a PocketBase app.
// This method would be implemented when we have access to the PocketBase API.
func (p *CortexPlugin) RegisterWithPocketBase(app interface{}) error {
	// This is a stub that would be implemented with the actual PocketBase API
	p.logger.Printf("Registering Cortex plugin with PocketBase")

	// If we have a custom HTTP handler that supports custom routes:
	if routeAdder, ok := app.(interface {
		OnBeforeServe() interface{}
	}); ok {
		p.logger.Printf("Found OnBeforeServe method in PocketBase app")

		// Get the router from PocketBase's OnBeforeServe hook
		e := routeAdder.OnBeforeServe()

		// If the router supports route registration methods:
		if router, ok := e.(interface {
			GET(path string, handler interface{}, middlewares ...interface{}) interface{}
			POST(path string, handler interface{}, middlewares ...interface{}) interface{}
			ANY(path string, handler interface{}, middlewares ...interface{}) interface{}
			Use(middlewares ...interface{}) interface{}
		}); ok {
			p.logger.Printf("Found router methods in PocketBase app")

			// Add middleware to preserve SSE headers
			router.Use(p.preserveSSEHeaders)

			// Register the routes
			basePath := strings.TrimSuffix(p.basePath, "/")

			// Register the SSE endpoint with GET method
			// IMPORTANT: Use a custom handler that guarantees correct Content-Type
			sseEndpoint := basePath + "/sse"
			p.logger.Printf("Registering SSE endpoint at %s", sseEndpoint)
			router.GET(sseEndpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Clear all headers to prevent any middleware from adding them
				for k := range w.Header() {
					w.Header().Del(k)
				}

				// Set SSE headers in the exact required order
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", "*")

				// Verify headers were set correctly
				contentType := w.Header().Get("Content-Type")
				if contentType != "text/event-stream" {
					p.logger.Printf("ERROR: Content-Type not set correctly! Got: %s", contentType)
					http.Error(w, "Server configuration error - invalid content type: "+contentType, http.StatusInternalServerError)
					return
				}

				// Delegate to the SSE handler
				p.GetSSEHandler().ServeHTTP(w, r)
			}))

			// Register the streamableHttp endpoint
			httpEndpoint := basePath + "/streamableHttp"
			p.logger.Printf("Registering streamableHttp endpoint at %s", httpEndpoint)

			// Register a dedicated handler that processes both GET and POST
			router.ANY(httpEndpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.logger.Printf("streamableHttp direct handler: %s %s", r.Method, r.URL.Path)

				// Clear all headers to prevent any middleware from adding them
				for k := range w.Header() {
					w.Header().Del(k)
				}

				// Set proper headers for all responses
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

				// Handle different HTTP methods
				switch r.Method {
				case http.MethodOptions:
					// For OPTIONS requests (CORS preflight), just return OK
					w.WriteHeader(http.StatusOK)
					return

				case http.MethodGet:
					// For GET requests, return server info in JSON-RPC 2.0 format
					serverInfo := p.GetServerInfo()
					response := map[string]interface{}{
						"jsonrpc": "2.0",
						"result": map[string]interface{}{
							"name":    serverInfo.Name,
							"version": serverInfo.Version,
							"status":  "ready",
							"endpoints": map[string]string{
								"sse":            p.basePath + "/sse",
								"message":        p.basePath + "/message",
								"tools":          p.basePath + "/tools",
								"streamableHttp": p.basePath + "/streamableHttp",
							},
						},
						"id": "server.info",
					}
					json.NewEncoder(w).Encode(response)
					return

				case http.MethodPost:
					// For POST requests, process JSON-RPC messages
					body, err := io.ReadAll(r.Body)
					if err != nil {
						p.logger.Printf("Error reading request body: %v", err)
						sendJSONRPCError(w, nil, -32700, "Error reading request body")
						return
					}

					// Only process if there's actual content
					if len(body) > 0 {
						var request map[string]interface{}
						if err := json.Unmarshal(body, &request); err != nil {
							p.logger.Printf("Error parsing JSON: %v", err)
							sendJSONRPCError(w, nil, -32700, "Parse error")
							return
						}

						// Extract method and ID
						method, _ := request["method"].(string)
						id := request["id"]

						// Check JSONRPC version
						version, _ := request["jsonrpc"].(string)
						if version != "2.0" {
							sendJSONRPCError(w, id, -32600, "Invalid Request: only JSON-RPC 2.0 is supported")
							return
						}

						// Handle the request using our shared method
						if method != "" {
							p.handleJSONRPCRequest(w, r, method, id, request)
							return
						} else {
							// For invalid/incomplete requests
							sendJSONRPCError(w, id, -32600, "Invalid Request: missing method")
							return
						}
					} else {
						// Empty POST
						sendJSONRPCError(w, nil, -32700, "Empty request")
						return
					}

				default:
					// Method not allowed
					sendJSONRPCError(w, nil, -32600, "Method not allowed")
					return
				}
			}))

			// Register the message endpoint with POST method
			messageEndpoint := basePath + "/message"
			p.logger.Printf("Registering message endpoint at %s", messageEndpoint)
			router.POST(messageEndpoint, p.GetHTTPHandler())

			return nil
		}
	}

	p.logger.Printf("WARNING: Could not register Cortex plugin with PocketBase. Manual integration required.")
	return nil
}

// GetServerInfo returns basic information about the Cortex server.
func (p *CortexPlugin) GetServerInfo() server.ServerInfo {
	return p.mcpServer.GetServerInfo()
}

// GetSSEHandler returns a dedicated handler just for the SSE endpoint
func (p *CortexPlugin) GetSSEHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.logger.Printf("SSE handler called: %s %s", r.Method, r.URL.Path)

		// Only support GET method for SSE
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// CRITICAL: We MUST clear all existing headers completely
		for k := range w.Header() {
			w.Header().Del(k)
		}

		// Set SSE headers - in this exact order
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Verify headers were set correctly
		if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
			p.logger.Printf("ERROR: Content-Type not set correctly! Got: %s", ct)
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Create a flusher for SSE
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Get or generate a session ID
		sessionID := r.URL.Query().Get("session")
		if sessionID == "" {
			sessionID = generateSessionID()
		}
		p.logger.Printf("SSE session established: %s", sessionID)

		// Get user agent for session tracking
		userAgent := r.UserAgent()

		// Create a context that cancels when the client disconnects
		sessionCtx, sessionCancel := context.WithCancel(r.Context())
		defer sessionCancel()

		// Create event queue for sending messages
		eventQueue := make(chan string, 100)
		defer close(eventQueue)

		// Register this session with the MCP server
		err := p.mcpServer.RegisterSession(sessionID, userAgent, func(msg []byte) error {
			select {
			case eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", msg):
				return nil
			case <-sessionCtx.Done():
				return fmt.Errorf("session closed")
			default:
				return fmt.Errorf("event queue full for session %s", sessionID)
			}
		})
		if err != nil {
			p.logger.Printf("Error registering session: %v", err)
			http.Error(w, "Could not register session", http.StatusInternalServerError)
			return
		}
		defer p.mcpServer.UnregisterSession(sessionID)

		// The message endpoint
		messageEndpoint := fmt.Sprintf("%s/message?sessionId=%s", p.basePath, sessionID)

		// Send the connected event - IMPORTANT: Must match exact Cortex server format
		fmt.Fprintf(w, "event: connected\ndata: {\"sessionId\": \"%s\"}\n\n", sessionID)
		flusher.Flush()

		// Send the endpoint event - IMPORTANT: Must match exact Cortex server format
		// Don't include quotes around the endpoint to prevent malformed URLs
		fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageEndpoint)
		flusher.Flush()

		// DO NOT send any server info message here - client will request when needed

		// Set up a heartbeat to keep the connection alive
		heartbeatTicker := time.NewTicker(30 * time.Second)
		defer heartbeatTicker.Stop()

		// Process events until the client disconnects
		for {
			select {
			case event := <-eventQueue:
				fmt.Fprint(w, event)
				flusher.Flush()
			case <-heartbeatTicker.C:
				// Send a heartbeat to keep the connection alive
				fmt.Fprint(w, ":\n\n") // Comment line in SSE spec
				flusher.Flush()
			case <-r.Context().Done():
				// Client disconnected
				p.logger.Printf("SSE session closed (client disconnected): %s", sessionID)
				return
			case <-sessionCtx.Done():
				// Session context done (canceled explicitly)
				return
			}
		}
	})
}

// GetHTTPHandler returns an HTTP handler for the Cortex server.
// This can be used to integrate with PocketBase or any other HTTP server.
func (p *CortexPlugin) GetHTTPHandler() http.Handler {
	// Create a handler that directly implements MCP protocol
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming request
		p.logger.Printf("MCP request received: %s %s", r.Method, r.URL.Path)

		// Set CORS headers for all responses (necessary for cross-origin requests)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cache-Control, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests - return immediately
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Extract the relative path (removing the base path)
		basePath := strings.TrimSuffix(p.basePath, "/")
		relPath := strings.TrimPrefix(r.URL.Path, basePath)

		// Ensure the path starts with a slash
		if relPath == "" {
			relPath = "/"
		} else if !strings.HasPrefix(relPath, "/") {
			relPath = "/" + relPath
		}

		p.logger.Printf("Rewriting path from %s to %s", r.URL.Path, relPath)

		// Handle direct streamableHttp connection attempts
		// This is typically the first connection the client tries before falling back to SSE
		if relPath == "/streamableHttp" || strings.HasSuffix(relPath, "/streamableHttp") {
			p.logger.Printf("Handling streamableHttp request at %s", r.URL.Path)

			// Clear all headers to prevent any middleware from adding them
			for k := range w.Header() {
				w.Header().Del(k)
			}

			// Set proper headers
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// Handle different HTTP methods
			switch r.Method {
			case http.MethodOptions:
				// For OPTIONS requests (CORS preflight), just return OK
				w.WriteHeader(http.StatusOK)
				return

			case http.MethodGet:
				// For GET requests, return server info in JSON-RPC 2.0 format
				serverInfo := p.GetServerInfo()
				response := map[string]interface{}{
					"jsonrpc": "2.0",
					"result": map[string]interface{}{
						"name":    serverInfo.Name,
						"version": serverInfo.Version,
						"status":  "ready",
						"endpoints": map[string]string{
							"sse":            p.basePath + "/sse",
							"message":        p.basePath + "/message",
							"tools":          p.basePath + "/tools",
							"streamableHttp": p.basePath + "/streamableHttp",
						},
					},
					"id": "server.info",
				}
				json.NewEncoder(w).Encode(response)
				return

			case http.MethodPost:
				// For POST requests, process JSON-RPC messages
				body, err := io.ReadAll(r.Body)
				if err != nil {
					p.logger.Printf("Error reading request body: %v", err)
					sendJSONRPCError(w, nil, -32700, "Error reading request body")
					return
				}

				// Only process if there's actual content
				if len(body) > 0 {
					var request map[string]interface{}
					if err := json.Unmarshal(body, &request); err != nil {
						p.logger.Printf("Error parsing JSON: %v", err)
						sendJSONRPCError(w, nil, -32700, "Parse error")
						return
					}

					// Extract method and ID
					method, _ := request["method"].(string)
					id := request["id"]

					// Check JSONRPC version
					version, _ := request["jsonrpc"].(string)
					if version != "2.0" {
						sendJSONRPCError(w, id, -32600, "Invalid Request: only JSON-RPC 2.0 is supported")
						return
					}

					// Handle the request using our shared method
					if method != "" {
						p.handleJSONRPCRequest(w, r, method, id, request)
						return
					} else {
						// For invalid/incomplete requests
						sendJSONRPCError(w, id, -32600, "Invalid Request: missing method")
						return
					}
				} else {
					// Empty POST
					sendJSONRPCError(w, nil, -32700, "Empty request")
					return
				}

			default:
				// Method not allowed
				sendJSONRPCError(w, nil, -32600, "Method not allowed")
				return
			}
		}

		// Handle specific MCP endpoints directly instead of using the embeddable server
		switch {
		case strings.HasSuffix(relPath, "/sse"):
			// Ensure headers are set correctly before delegating to the SSE handler
			// First, clear existing headers that might interfere
			for k := range w.Header() {
				w.Header().Del(k)
			}

			// Set required SSE headers in correct order
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// Double-check Content-Type before proceeding
			if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
				p.logger.Printf("ERROR: Content-Type not set correctly before SSE handler! Got: %s", ct)
			}

			// Now delegate to the SSE handler
			p.GetSSEHandler().ServeHTTP(w, r)
			return

		case strings.HasSuffix(relPath, "/message"):
			// Handle message endpoint (JSON-RPC)
			p.logger.Printf("Message endpoint detected")

			// Only accept POST requests for messages
			if r.Method != http.MethodPost {
				sendJSONRPCError(w, nil, -32600, "Method not allowed")
				return
			}

			// Extract sessionId from query parameters
			sessionID := r.URL.Query().Get("sessionId")
			if sessionID == "" {
				sendJSONRPCError(w, nil, -32602, "Missing sessionId parameter")
				return
			}

			// Read request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				sendJSONRPCError(w, nil, -32700, "Error reading request body")
				return
			}

			// Parse JSON-RPC request
			var request map[string]interface{}
			if err := json.Unmarshal(body, &request); err != nil {
				sendJSONRPCError(w, nil, -32700, "Parse error")
				return
			}

			// Check JSONRPC version
			version, _ := request["jsonrpc"].(string)
			if version != "2.0" {
				sendJSONRPCError(w, nil, -32600, "Invalid Request: only JSON-RPC 2.0 is supported")
				return
			}

			// Extract method and ID
			method, _ := request["method"].(string)
			id := request["id"]

			// Handle the request using our shared method
			p.handleJSONRPCRequest(w, r, method, id, request)
			return

		default:
			// Handle other endpoints
			w.Header().Set("Content-Type", "application/json")

			// For /tools endpoint, return tools list
			if relPath == "/tools" || relPath == "/tools/" {
				w.Header().Set("Content-Type", "application/json")
				tools := make([]map[string]interface{}, 0)

				// Log all tools to see their content
				p.logger.Printf("==== DEBUG: All registered tools ====")
				allTools := p.mcpServer.GetTools()
				for name, tool := range allTools {
					toolBytes, _ := json.Marshal(tool)
					p.logger.Printf("Tool %s: %s", name, string(toolBytes))
				}

				// Convert registered tools to the expected format
				for name, tool := range p.mcpServer.GetTools() {
					// Log the tool details to debug
					toolBytes, _ := json.Marshal(tool)
					p.logger.Printf("Processing tool: %s, Details: %s", name, string(toolBytes))

					toolInfo := map[string]interface{}{
						"name":        name,
						"description": tool.Description,
					}

					// Add parameters
					params := make([]map[string]interface{}, 0, len(tool.Parameters))
					for _, param := range tool.Parameters {
						paramInfo := map[string]interface{}{
							"name":        param.Name,
							"description": param.Description,
							"type":        param.Type,
							"required":    param.Required,
						}
						if param.Items != nil {
							paramInfo["items"] = param.Items
						}
						params = append(params, paramInfo)
					}

					toolInfo["parameters"] = params

					// Use the tool's inputSchema if available, otherwise create one
					if tool.InputSchema != nil {
						toolInfo["inputSchema"] = tool.InputSchema
						p.logger.Printf("Using existing inputSchema for tool %s", name)
					} else {
						// Add inputSchema manually
						p.logger.Printf("Generating inputSchema for tool %s", name)
						properties := make(map[string]interface{})
						required := []string{}

						for _, param := range tool.Parameters {
							paramSchema := map[string]interface{}{
								"type":        param.Type,
								"description": param.Description,
							}

							if param.Items != nil {
								paramSchema["items"] = param.Items
							}

							properties[param.Name] = paramSchema

							if param.Required {
								required = append(required, param.Name)
							}
						}

						toolInfo["inputSchema"] = map[string]interface{}{
							"type":       "object",
							"properties": properties,
							"required":   required,
						}
					}

					tools = append(tools, toolInfo)

					// Log the final tool info for debugging
					infoBytes, _ := json.Marshal(toolInfo)
					p.logger.Printf("Final tool info: %s", string(infoBytes))
				}

				// Create the response with a fixed ID for the /tools endpoint
				responseObj := map[string]interface{}{
					"jsonrpc": "2.0",
					"result": map[string]interface{}{
						"tools": tools,
					},
					"id": "tools.list",
				}

				// Log the final response for debugging
				respBytes, _ := json.Marshal(responseObj)
				p.logger.Printf("tools/list response: %s", string(respBytes))

				// Return the tools list in JSON-RPC 2.0 format
				json.NewEncoder(w).Encode(responseObj)
				return
			}

			// For the root endpoint, return simple server info
			if relPath == "/" {
				w.Header().Set("Content-Type", "application/json")
				serverInfo := p.GetServerInfo()

				// Use JSON-RPC 2.0 format for server info to be compatible with client expectations
				response := map[string]interface{}{
					"jsonrpc": "2.0",
					"result": map[string]interface{}{
						"name":    serverInfo.Name,
						"version": serverInfo.Version,
						"status":  "ready",
						"endpoints": map[string]string{
							"sse":            p.basePath + "/sse",
							"message":        p.basePath + "/message",
							"tools":          p.basePath + "/tools",
							"streamableHttp": p.basePath + "/streamableHttp",
						},
					},
					"id": "server.info",
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			// Default to 404 for unknown endpoints
			http.NotFound(w, r)
		}
	})

	return handler
}

// GetBasePath returns the base path configured for the plugin.
func (p *CortexPlugin) GetBasePath() string {
	return p.basePath
}

// Helper function to generate a session ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

// sendJSONRPCError sends a JSON-RPC 2.0 error response
func sendJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")

	// JSON-RPC 2.0 error response structure
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	// Only include id if it's not nil
	if id != nil {
		response["id"] = id
	} else {
		response["id"] = nil // Explicitly set to null in JSON
	}

	// Marshal to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		// If we can't marshal the error, send a simple error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal JSON-RPC error"},"id":null}`))
		return
	}

	// Send the response
	w.WriteHeader(http.StatusOK) // Always 200 OK for JSON-RPC, error is in the response
	w.Write(responseBytes)
}

// preserveSSEHeaders creates a middleware that preserves SSE headers
// by clearing any existing headers and setting the correct ones
func (p *CortexPlugin) preserveSSEHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer that prevents header modifications
		if strings.HasSuffix(r.URL.Path, "/sse") {
			p.logger.Printf("Preserving SSE headers for %s", r.URL.Path)

			// Use custom response writer for SSE endpoints only
			next.ServeHTTP(&preservingResponseWriter{
				ResponseWriter: w,
				logger:         p.logger,
			}, r)
		} else {
			// For other endpoints, use normal handler
			next.ServeHTTP(w, r)
		}
	})
}

// preservingResponseWriter ensures headers can't be modified after they're set
type preservingResponseWriter struct {
	http.ResponseWriter
	headerWritten bool
	logger        *log.Logger
}

// Header returns the header map to be sent
func (w *preservingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader sends an HTTP response header
func (w *preservingResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}

	// Ensure Content-Type is correct for SSE
	if w.Header().Get("Content-Type") != "text/event-stream" {
		w.Header().Set("Content-Type", "text/event-stream")
		w.logger.Printf("Forcing Content-Type to text/event-stream")
	}

	w.ResponseWriter.WriteHeader(statusCode)
	w.headerWritten = true
}

// Write writes the data to the connection
func (w *preservingResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// handleJSONRPCRequest processes a JSON-RPC request and responds with the same ID
func (p *CortexPlugin) handleJSONRPCRequest(w http.ResponseWriter, r *http.Request, method string, id interface{}, request map[string]interface{}) {
	// Get session ID from query parameters if available
	sessionID := r.URL.Query().Get("sessionId")

	// Create a context that may include the session ID
	ctx := r.Context()
	if sessionID != "" {
		ctx = context.WithValue(ctx, "sessionId", sessionID)
	}

	var response map[string]interface{}

	// Process the request based on method
	switch method {
	case "initialize":
		// Return server information
		serverInfo := p.GetServerInfo()
		response = map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]interface{}{
					"name":    serverInfo.Name,
					"version": serverInfo.Version,
				},
				"capabilities": map[string]interface{}{
					"resources": map[string]bool{
						"listChanged": true,
					},
					"tools": map[string]bool{
						"listChanged": true,
					},
					"prompts": map[string]bool{
						"listChanged": true,
					},
					"logging": struct{}{},
				},
			},
			"id": id,
		}

	case "tools/list":
		// Get tools from the MCP server
		tools := make([]map[string]interface{}, 0)

		// Log all tools to see their content
		p.logger.Printf("==== DEBUG: All registered tools ====")
		allTools := p.mcpServer.GetTools()
		for name, tool := range allTools {
			toolBytes, _ := json.Marshal(tool)
			p.logger.Printf("Tool %s: %s", name, string(toolBytes))
		}

		// Convert registered tools to the expected format
		for name, tool := range p.mcpServer.GetTools() {
			// Log the tool details to debug
			toolBytes, _ := json.Marshal(tool)
			p.logger.Printf("Processing tool: %s, Details: %s", name, string(toolBytes))

			toolInfo := map[string]interface{}{
				"name":        name,
				"description": tool.Description,
			}

			// Add parameters
			params := make([]map[string]interface{}, 0, len(tool.Parameters))
			for _, param := range tool.Parameters {
				paramInfo := map[string]interface{}{
					"name":        param.Name,
					"description": param.Description,
					"type":        param.Type,
					"required":    param.Required,
				}
				if param.Items != nil {
					paramInfo["items"] = param.Items
				}
				params = append(params, paramInfo)
			}

			toolInfo["parameters"] = params

			// Use the tool's inputSchema if available, otherwise create one
			if tool.InputSchema != nil {
				toolInfo["inputSchema"] = tool.InputSchema
				p.logger.Printf("Using existing inputSchema for tool %s", name)
			} else {
				// Add inputSchema manually
				p.logger.Printf("Generating inputSchema for tool %s", name)
				properties := make(map[string]interface{})
				required := []string{}

				for _, param := range tool.Parameters {
					paramSchema := map[string]interface{}{
						"type":        param.Type,
						"description": param.Description,
					}

					if param.Items != nil {
						paramSchema["items"] = param.Items
					}

					properties[param.Name] = paramSchema

					if param.Required {
						required = append(required, param.Name)
					}
				}

				toolInfo["inputSchema"] = map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				}
			}

			tools = append(tools, toolInfo)

			// Log the final tool info for debugging
			infoBytes, _ := json.Marshal(toolInfo)
			p.logger.Printf("Final tool info: %s", string(infoBytes))
		}

		// Create the response with the same ID as the request
		response = map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"tools": tools,
			},
			"id": id,
		}

		// Log the final response for debugging
		respBytes, _ := json.Marshal(response)
		p.logger.Printf("tools/list response: %s", string(respBytes))

	default:
		// Handle tool execution for methods with tools/ prefix
		if strings.HasPrefix(method, "tools/") {
			// Create client session
			clientSession := &types.ClientSession{
				ID:        sessionID,
				UserAgent: r.UserAgent(),
				Connected: true,
			}

			// Extract parameters
			params, _ := request["params"].(map[string]interface{})
			if params == nil {
				params = make(map[string]interface{})
			}

			// Execute the tool through the MCP server
			toolName := strings.TrimPrefix(method, "tools/")
			result, err := p.mcpServer.ExecuteTool(ctx, server.ToolCallRequest{
				Name:       toolName,
				Parameters: params,
				Session:    clientSession,
			})

			if err != nil {
				// Create error response with the same ID
				sendJSONRPCError(w, id, -32603, err.Error())
				return
			}

			// Create success response with the same ID
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  result,
				"id":      id,
			}
		} else {
			// Method not found - respond with the same ID
			sendJSONRPCError(w, id, -32601, fmt.Sprintf("Method not found: %s", method))
			return
		}
	}

	// Send response
	if response != nil {
		// Marshal response
		respBytes, err := json.Marshal(response)
		if err != nil {
			sendJSONRPCError(w, id, -32603, "Internal error serializing response")
			return
		}

		// Send through both SSE channel and HTTP response if session exists
		if sessionID != "" {
			err = p.mcpServer.SendToSession(sessionID, respBytes)
			if err != nil {
				p.logger.Printf("Warning: Could not send response via SSE: %v", err)
			}
		}

		// Write HTTP response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}
}
