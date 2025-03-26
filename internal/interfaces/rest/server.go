// Package rest provides the HTTP interface for the MCP server.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

const (
	// JSON-RPC version used by the MCP protocol
	jsonRPCVersion = "2.0"

	// MCP protocol version
	mcpProtocolVersion = "2024-11-05"

	// Default notification buffer size
	defaultNotificationBufferSize = 100
)

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPServer represents the HTTP server for the MCP protocol.
type MCPServer struct {
	service     *usecases.ServerService
	httpServer  *http.Server
	sseClients  sync.Map
	notifier    *server.NotificationSender
	sseServer   *server.SSEServer
	sessionRepo domain.SessionRepository
}

// NewMCPServer creates a new MCP server.
func NewMCPServer(service *usecases.ServerService, addr string) *MCPServer {
	notifier := server.NewNotificationSender(jsonRPCVersion)

	s := &MCPServer{
		service:  service,
		notifier: notifier,
	}

	// Create SSE Server with MCP message handler
	sseServer := server.NewSSEServer(notifier, s.handleMessage,
		server.WithMessageEndpoint("/message"),
		server.WithSSEEndpoint("/sse"),
		server.WithBasePath(""),
	)

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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
	log.Printf("Starting MCP server on %s", s.httpServer.Addr)
	log.Printf("Available endpoints: /, /jsonrpc, /sse, /message, /events, /status")
	return s.httpServer.ListenAndServe()
}

// Stop stops the MCP server.
func (s *MCPServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// handleMessage processes a JSON-RPC message and returns a response.
// This is used by the SSE server to handle incoming messages.
func (s *MCPServer) handleMessage(ctx context.Context, rawMessage json.RawMessage) interface{} {
	// Parse JSON-RPC request
	var request JSONRPCRequest
	if err := json.Unmarshal(rawMessage, &request); err != nil {
		return createErrorResponse(nil, -32700, "Parse error")
	}

	// Validate JSON-RPC version
	if request.JSONRPC != jsonRPCVersion {
		return createErrorResponse(request.ID, -32600, "Invalid JSON-RPC version")
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
		return createErrorResponse(request.ID, -32601, fmt.Sprintf("Method '%s' not found", request.Method))
	}
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Process the message
	response := s.handleMessage(ctx, body)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods for processing specific JSON-RPC methods

func (s *MCPServer) processInitialize(ctx context.Context, request JSONRPCRequest) interface{} {
	// Log initialization request
	log.Printf("Processing initialize request")

	// Parse initialization parameters if needed
	// ...

	// Get server info
	name, version, instructions := s.service.ServerInfo()

	log.Printf("Server info: %s %s", name, version)

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

	log.Printf("Processed initialize response with protocol version %s", mcpProtocolVersion)
	return createResponse(request.ID, result)
}

func (s *MCPServer) processPing(ctx context.Context, request JSONRPCRequest) interface{} {
	// Simple ping response
	log.Printf("Processing ping request")
	return createResponse(request.ID, struct{}{})
}

func (s *MCPServer) processResourcesList(ctx context.Context, request JSONRPCRequest) interface{} {
	log.Printf("Processing resources/list request")
	resources, err := s.service.ListResources(ctx)
	if err != nil {
		log.Printf("Error listing resources: %v", err)
		return createErrorResponse(request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
	}

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

	log.Printf("Processed resources/list response with %d resources", len(resources))
	return createResponse(request.ID, result)
}

func (s *MCPServer) processResourcesRead(ctx context.Context, request JSONRPCRequest) interface{} {
	log.Printf("Processing resources/read request")

	// Extract URI from parameters
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		log.Printf("Invalid params, expected map")
		return createErrorResponse(request.ID, -32602, "Invalid params")
	}

	uri, ok := params["uri"].(string)
	if !ok || uri == "" {
		log.Printf("Missing or invalid 'uri' parameter")
		return createErrorResponse(request.ID, -32602, "Missing or invalid 'uri' parameter")
	}

	log.Printf("Reading resource with URI: %s", uri)

	// Get resource
	resource, err := s.service.GetResource(ctx, uri)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("Resource not found: %s", uri)
			return createErrorResponse(request.ID, 404, fmt.Sprintf("Resource not found: %s", uri))
		} else {
			log.Printf("Error getting resource: %v", err)
			return createErrorResponse(request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
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

	log.Printf("Processed resources/read response for URI: %s", uri)
	return createResponse(request.ID, result)
}

func (s *MCPServer) processToolsList(ctx context.Context, request JSONRPCRequest) interface{} {
	log.Printf("Processing tools/list request")
	tools, err := s.service.ListTools(ctx)
	if err != nil {
		log.Printf("Error listing tools: %v", err)
		return createErrorResponse(request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
	}

	// Convert domain tools to response format
	toolList := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
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

	log.Printf("Processed tools/list response with %d tools", len(tools))
	return createResponse(request.ID, result)
}

func (s *MCPServer) processToolsCall(ctx context.Context, request JSONRPCRequest) interface{} {
	log.Printf("Processing tools/call request: %+v", request)

	// Extract parameters
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		log.Printf("Invalid params, expected map: %T", request.Params)
		return createErrorResponse(request.ID, -32602, "Invalid params")
	}

	// Get tool name
	toolName, ok := params["name"].(string)
	if !ok || toolName == "" {
		log.Printf("Missing or invalid 'name' parameter")
		return createErrorResponse(request.ID, -32602, "Missing or invalid 'name' parameter")
	}

	// Get tool parameters
	toolParams, ok := params["parameters"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid or missing 'parameters' field")
		toolParams = map[string]interface{}{}
	}

	log.Printf("Tool call request for tool '%s' with params: %+v", toolName, toolParams)

	// Get the tool
	_, err := s.service.GetTool(ctx, toolName)
	if err != nil {
		log.Printf("Error getting tool '%s': %v", toolName, err)
		if errors.Is(err, domain.ErrNotFound) {
			return createErrorResponse(request.ID, 404, fmt.Sprintf("Tool not found: %s", toolName))
		} else {
			return createErrorResponse(request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
		}
	}

	// Handle specific tools
	var result interface{}

	switch toolName {
	case "echo":
		// Handle echo tool
		message, ok := toolParams["message"].(string)
		if !ok || message == "" {
			log.Printf("Missing or invalid 'message' parameter for echo tool")
			return createErrorResponse(request.ID, -32602, "Missing or invalid 'message' parameter")
		}

		// Echo the message back
		result = map[string]interface{}{
			"message": message,
		}

	default:
		log.Printf("Tool '%s' exists but handler not implemented", toolName)
		return createErrorResponse(request.ID, -32603, fmt.Sprintf("Tool handler not implemented for: %s", toolName))
	}

	// Return response
	resp := map[string]interface{}{
		"result": result,
	}

	log.Printf("Processed tools/call response for tool '%s'", toolName)
	return createResponse(request.ID, resp)
}

func (s *MCPServer) processPromptsList(ctx context.Context, request JSONRPCRequest) interface{} {
	log.Printf("Processing prompts/list request")
	prompts, err := s.service.ListPrompts(ctx)
	if err != nil {
		log.Printf("Error listing prompts: %v", err)
		return createErrorResponse(request.ID, -32603, fmt.Sprintf("Internal error: %v", err))
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

	log.Printf("Processed prompts/list response with %d prompts", len(prompts))
	return createResponse(request.ID, result)
}

func (s *MCPServer) processPromptsGet(ctx context.Context, request JSONRPCRequest) interface{} {
	log.Printf("Processing prompts/get request")
	// TODO: Implement prompt get handler
	return createErrorResponse(request.ID, -32603, "Prompt get not implemented")
}

// Helper functions for creating JSON-RPC responses

func createResponse(id interface{}, result interface{}) interface{} {
	return JSONRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func createErrorResponse(id interface{}, code int, message string) interface{} {
	return JSONRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}

// For backward compatibility - to be removed after transitioning to the new SSE server
func sendJSONRPCResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	response := JSONRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// For backward compatibility - to be removed after transitioning to the new SSE server
func sendJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := JSONRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetServerInfo returns information about the server.
// This is useful for external components that need access to the server information.
func (s *MCPServer) GetServerInfo() (name string, version string, instructions string) {
	return s.service.ServerInfo()
}
