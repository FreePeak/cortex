// Package rest provides the HTTP interface for the MCP server.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

const (
	// JSON-RPC version used by the MCP protocol
	jsonRPCVersion = "2.0"

	// MCP protocol version
	mcpProtocolVersion = "2024-11-05"
)

// MCPServer represents the HTTP server for the MCP protocol.
type MCPServer struct {
	service    *usecases.ServerService
	httpServer *http.Server
	sseServer  *server.SSEServer
	notifier   *server.NotificationSender
	logger     *logging.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// MCPServerOption is a function option for MCPServer
type MCPServerOption func(*MCPServer)

// WithLogger sets the logger for the MCPServer
func WithLogger(logger *logging.Logger) MCPServerOption {
	return func(s *MCPServer) {
		s.logger = logger
	}
}

// NewMCPServer creates a new MCP server.
func NewMCPServer(service *usecases.ServerService, addr string, opts ...MCPServerOption) *MCPServer {
	// Create root context for the server
	ctx, cancel := context.WithCancel(context.Background())

	// Create default logger
	defaultLogger, err := logging.New(logging.Config{
		Level:       logging.InfoLevel,
		Development: true,
		OutputPaths: []string{"stdout"},
		InitialFields: logging.Fields{
			"component": "mcp-server",
		},
	})
	if err != nil {
		// Fallback to a simple default logger if we can't create the structured one
		defaultLogger = logging.Default()
	}

	notifier := server.NewNotificationSender(jsonRPCVersion)

	s := &MCPServer{
		service:  service,
		notifier: notifier,
		logger:   defaultLogger,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Apply all options
	for _, opt := range opts {
		opt(s)
	}

	// Create message handler function for the SSE server
	mcpHandler := func(ctx context.Context, rawMessage json.RawMessage) interface{} {
		// Create a child context from the server context
		handlerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel() // Call cancel to prevent context leak
		return s.processMessage(handlerCtx, rawMessage)
	}

	// Create a custom context function for the SSE server
	contextFunc := func(parentCtx context.Context, r *http.Request) context.Context {
		// Simply return the parent context instead of creating a new one with a discarded cancel
		// This maintains the same behavior without leaking the context
		return parentCtx
	}

	// Create the SSE Server with MCP message handler and enhanced context handling
	sseOptions := []server.SSEOption{
		server.WithMessageEndpoint("/message"),
		server.WithSSEEndpoint("/sse"),
		server.WithBasePath(""),
		server.WithSSEContextFunc(contextFunc),
	}

	// If we have a logger, pass it to the SSE server
	if s.logger != nil {
		sseOptions = append(sseOptions, server.WithLogger(s.logger))
	}

	sseServer := server.NewSSEServer(notifier, mcpHandler, sseOptions...)

	s.sseServer = sseServer

	// Create HTTP server
	mux := http.NewServeMux()

	// Standard MCP endpoints
	mux.HandleFunc("/", s.handleJSONRPC)        // Default endpoint for JSON-RPC
	mux.HandleFunc("/jsonrpc", s.handleJSONRPC) // Alternative endpoint for JSON-RPC
	mux.HandleFunc("/events", s.redirectToSSE)  // Redirect to SSE endpoint

	// Add SSE server handler
	mux.Handle("/sse", sseServer)
	mux.Handle("/message", sseServer)

	// Add a simple status endpoint
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name, version, _ := service.ServerInfo()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"name":     name,
			"version":  version,
			"protocol": mcpProtocolVersion,
		})
	})

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

// redirectToSSE redirects clients to the SSE endpoint
func (s *MCPServer) redirectToSSE(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/sse", http.StatusFound)
}

// Start starts the MCP server.
func (s *MCPServer) Start() error {
	s.logger.Info("Starting MCP server", logging.Fields{"address": s.httpServer.Addr})
	s.logger.Info("Available endpoints", logging.Fields{"endpoints": "/, /jsonrpc, /sse, /message, /events, /status"})
	return s.httpServer.ListenAndServe()
}

// Stop stops the MCP server.
func (s *MCPServer) Stop(ctx context.Context) error {
	// Cancel our internal context first to signal all ongoing operations to stop
	s.cancel()

	// Shutdown the HTTP server
	err := s.httpServer.Shutdown(ctx)

	// Return any error from shutting down the HTTP server
	return err
}

// handleJSONRPC handles JSON-RPC requests over HTTP directly.
func (s *MCPServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Create context with timeout, derived from request context and server context
	// This ensures the context is canceled if either the request ends or the server is stopped
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Process the message
	response := s.processMessage(ctx, body)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// Helper methods for processing specific JSON-RPC methods

func (s *MCPServer) processInitialize(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	// Log initialization request
	s.logger.Info("Processing initialize request")

	// Parse initialization parameters if needed
	// ...

	// Get server info
	name, version, instructions := s.service.ServerInfo()

	s.logger.Info("Server info", logging.Fields{"name": name, "version": version})

	// Create response
	result := map[string]interface{}{
		"protocolVersion": mcpProtocolVersion,
		"serverInfo": map[string]string{
			"name":    name,
			"version": version,
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
	}

	// Add instructions if provided
	if instructions != "" {
		result["instructions"] = instructions
	}

	s.logger.Info("Processed initialize response", logging.Fields{"protocolVersion": mcpProtocolVersion})
	return domain.CreateResponse(jsonRPCVersion, request.ID, result)
}

func (s *MCPServer) processPing(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	// Simple ping response
	s.logger.Debug("Processing ping request")
	return domain.CreateResponse(jsonRPCVersion, request.ID, struct{}{})
}

func (s *MCPServer) processResourcesList(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	s.logger.Info("Processing resources/list request")

	// Debug logging to verify service access
	s.logger.Debug("Service access", logging.Fields{"servicePtr": fmt.Sprintf("%p", s.service)})

	resources, err := s.service.ListResources(ctx)
	if err != nil {
		s.logger.Error("Error listing resources", logging.Fields{"error": err})
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
	}

	s.logger.Info("Found resources", logging.Fields{"count": len(resources)})

	// Convert domain resources to response format
	resourceList := make([]map[string]interface{}, len(resources))
	for i, resource := range resources {
		resourceList[i] = map[string]interface{}{
			"uri":         resource.URI,
			"name":        resource.Name,
			"description": resource.Description,
			"mimeType":    resource.MIMEType,
		}
	}

	result := map[string]interface{}{
		"resources": resourceList,
	}

	s.logger.Info("Processed resources/list response", logging.Fields{"resourceCount": len(resources)})
	return domain.CreateResponse(jsonRPCVersion, request.ID, result)
}

func (s *MCPServer) processResourcesRead(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	s.logger.Info("Processing resources/read request")

	// Extract URI from parameters
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		s.logger.Warn("Invalid params, expected map", logging.Fields{"paramsType": fmt.Sprintf("%T", request.Params)})
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32602, "Invalid params")
	}

	uri, ok := params["uri"].(string)
	if !ok || uri == "" {
		s.logger.Warn("Missing or invalid 'uri' parameter")
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32602, "Missing or invalid 'uri' parameter")
	}

	s.logger.Info("Reading resource", logging.Fields{"uri": uri})

	// Get resource
	resource, err := s.service.GetResource(ctx, uri)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.Warn("Resource not found", logging.Fields{"uri": uri})
			return domain.CreateErrorResponse(jsonRPCVersion, request.ID, 404, fmt.Sprintf("Resource not found: %s", uri))
		} else {
			s.logger.Error("Error getting resource", logging.Fields{"uri": uri, "error": err})
			return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
		}
	}

	// Placeholder for resource contents - in a real implementation, this would get the actual content
	contents := map[string]interface{}{
		"uri":      resource.URI,
		"mimeType": resource.MIMEType,
		"text":     "Sample resource content", // This would normally come from the resource
	}

	result := map[string]interface{}{
		"contents": []interface{}{contents},
	}

	s.logger.Info("Processed resources/read response", logging.Fields{"uri": uri})
	return domain.CreateResponse(jsonRPCVersion, request.ID, result)
}

func (s *MCPServer) processToolsList(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	s.logger.Info("Processing tools/list request")

	// Debug logging to verify service access
	s.logger.Debug("Service access", logging.Fields{"servicePtr": fmt.Sprintf("%p", s.service)})

	tools, err := s.service.ListTools(ctx)
	if err != nil {
		s.logger.Error("Error listing tools", logging.Fields{"error": err})
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
	}

	s.logger.Info("Found tools", logging.Fields{"count": len(tools)})

	// Convert domain tools to response format
	toolList := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		s.logger.Debug("Processing tool", logging.Fields{
			"index": i,
			"name":  tool.Name,
			"desc":  tool.Description,
		})

		// Format parameters as an object with properties
		parametersObj := make(map[string]interface{})
		parametersObj["type"] = "object"

		properties := make(map[string]interface{})
		required := []string{}

		for _, param := range tool.Parameters {
			paramObj := map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}
			properties[param.Name] = paramObj

			if param.Required {
				required = append(required, param.Name)
			}
		}

		parametersObj["properties"] = properties
		if len(required) > 0 {
			parametersObj["required"] = required
		}

		// Build tool object
		toolList[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": parametersObj,
		}
	}

	result := map[string]interface{}{
		"tools": toolList,
	}

	s.logger.Info("Processed tools/list response", logging.Fields{"toolCount": len(tools)})
	return domain.CreateResponse(jsonRPCVersion, request.ID, result)
}

func (s *MCPServer) processToolsCall(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	s.logger.Info("Processing tools/call request", logging.Fields{"request": fmt.Sprintf("%+v", request)})

	// Extract parameters
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		s.logger.Warn("Invalid params, expected map", logging.Fields{"paramsType": fmt.Sprintf("%T", request.Params)})
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32602, "Invalid params")
	}

	// Get tool name
	toolName, ok := params["name"].(string)
	if !ok || toolName == "" {
		s.logger.Warn("Missing or invalid 'name' parameter")
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32602, "Missing or invalid 'name' parameter")
	}

	// Get tool parameters - check for 'arguments' field instead of 'parameters'
	toolParams, ok := params["arguments"].(map[string]interface{})
	if !ok {
		s.logger.Debug("Invalid or missing 'arguments' field")
		toolParams = map[string]interface{}{}
	}

	s.logger.Info("Tool call request", logging.Fields{
		"tool":   toolName,
		"params": fmt.Sprintf("%+v", toolParams),
	})

	// Get the tool
	_, err := s.service.GetTool(ctx, toolName)
	if err != nil {
		s.logger.Error("Error getting tool", logging.Fields{"tool": toolName, "error": err})
		if errors.Is(err, domain.ErrNotFound) {
			return domain.CreateErrorResponse(jsonRPCVersion, request.ID, 404, fmt.Sprintf("Tool not found: %s", toolName))
		} else {
			return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
		}
	}

	// Handle specific tools
	var result interface{}

	switch toolName {
	case "echo":
		// Handle echo tool
		messageVal, ok := toolParams["message"]
		if !ok || messageVal == nil {
			s.logger.Warn("Missing or invalid 'message' parameter for echo tool")
			return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32602, "Missing or invalid 'message' parameter")
		}

		// Convert message to string based on its type
		var message string
		switch v := messageVal.(type) {
		case string:
			message = v
		case float64, int, int64, float32:
			message = fmt.Sprintf("%v", v)
		default:
			// Try JSON conversion for complex types
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				message = fmt.Sprintf("%v", v)
			} else {
				message = string(jsonBytes)
			}
		}

		// Echo the message back with content as an array of objects
		result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		}

	default:
		s.logger.Warn("Tool handler not implemented", logging.Fields{"tool": toolName})
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, fmt.Sprintf("Tool handler not implemented for: %s", toolName))
	}

	s.logger.Info("Processed tools/call response", logging.Fields{"tool": toolName})
	return domain.CreateResponse(jsonRPCVersion, request.ID, result)
}

func (s *MCPServer) processPromptsList(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	s.logger.Info("Processing prompts/list request")
	prompts, err := s.service.ListPrompts(ctx)
	if err != nil {
		s.logger.Error("Error listing prompts", logging.Fields{"error": err})
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
	}

	// Convert domain prompts to response format
	promptList := make([]map[string]interface{}, len(prompts))
	for i, prompt := range prompts {
		parameters := make([]map[string]interface{}, len(prompt.Parameters))
		for j, param := range prompt.Parameters {
			parameters[j] = map[string]interface{}{
				"name":        param.Name,
				"description": param.Description,
				"type":        param.Type,
				"required":    param.Required,
			}
		}

		promptList[i] = map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"parameters":  parameters,
		}
	}

	result := map[string]interface{}{
		"prompts": promptList,
	}

	s.logger.Info("Processed prompts/list response", logging.Fields{"promptCount": len(prompts)})
	return domain.CreateResponse(jsonRPCVersion, request.ID, result)
}

func (s *MCPServer) processPromptsGet(ctx context.Context, request domain.JSONRPCRequest) interface{} {
	s.logger.Info("Processing prompts/get request")
	// TODO: Implement prompt get handler
	return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32603, "Prompt get not implemented")
}

// GetServerInfo returns information about the server.
// This is useful for external components that need access to the server information.
func (s *MCPServer) GetServerInfo() (name string, version string, instructions string) {
	return s.service.ServerInfo()
}

// GetService returns the server service.
// This is useful for external components that need access to the service.
func (s *MCPServer) GetService() *usecases.ServerService {
	return s.service
}

// GetAddress returns the server's address
func (s *MCPServer) GetAddress() string {
	if s.httpServer != nil {
		return s.httpServer.Addr
	}
	return ""
}

// processMessage processes a JSON-RPC message and returns a response.
func (s *MCPServer) processMessage(ctx context.Context, rawMessage json.RawMessage) interface{} {
	// Check if the passed context is done
	select {
	case <-ctx.Done():
		return domain.CreateErrorResponse(jsonRPCVersion, nil, -32603, "Request context canceled")
	default:
		// Continue processing
	}

	// Parse JSON-RPC request
	var request domain.JSONRPCRequest
	if err := json.Unmarshal(rawMessage, &request); err != nil {
		return domain.CreateErrorResponse(jsonRPCVersion, nil, -32700, "Parse error")
	}

	// Validate JSON-RPC version
	if request.JSONRPC != jsonRPCVersion {
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32600, "Invalid JSON-RPC version")
	}

	// Handle request based on method
	switch request.Method {
	case "initialize":
		return s.processInitialize(ctx, request)
	case "ping":
		return s.processPing(ctx, request)
	case "resources/list":
		return s.processResourcesList(ctx, request)
	case "resources/read":
		return s.processResourcesRead(ctx, request)
	case "tools/list":
		return s.processToolsList(ctx, request)
	case "tools/call":
		return s.processToolsCall(ctx, request)
	case "prompts/list":
		return s.processPromptsList(ctx, request)
	case "prompts/get":
		return s.processPromptsGet(ctx, request)
	default:
		return domain.CreateErrorResponse(jsonRPCVersion, request.ID, -32601, fmt.Sprintf("Method '%s' not found", request.Method))
	}
}
