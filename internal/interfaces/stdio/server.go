// Package stdio provides the stdio interface for the MCP server.
package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
)

// StdioContextFunc is a function that takes an existing context and returns
// a potentially modified context.
// This can be used to inject context values from environment variables,
// for example.
type StdioContextFunc func(ctx context.Context) context.Context

// StdioServer wraps a MCPServer and handles stdio communication.
// It provides a simple way to create command-line MCP servers that
// communicate via standard input/output streams using JSON-RPC messages.
type StdioServer struct {
	server      *rest.MCPServer
	errLogger   *log.Logger
	contextFunc StdioContextFunc
	mu          sync.Mutex
}

// StdioOption defines a function type for configuring StdioServer
type StdioOption func(*StdioServer)

// WithErrorLogger sets the error logger for the server
func WithErrorLogger(logger *log.Logger) StdioOption {
	return func(s *StdioServer) {
		s.errLogger = logger
	}
}

// WithContextFunc sets a function that will be called to customize the context
// to the server. Note that the stdio server uses the same context for all requests,
// so this function will only be called once per server instance.
func WithStdioContextFunc(fn StdioContextFunc) StdioOption {
	return func(s *StdioServer) {
		s.contextFunc = fn
	}
}

// NewStdioServer creates a new stdio server wrapper around an MCPServer.
// It initializes the server with a default error logger that logs to stderr.
func NewStdioServer(server *rest.MCPServer, opts ...StdioOption) *StdioServer {
	s := &StdioServer{
		server:    server,
		errLogger: log.New(os.Stderr, "[STDIO] ", log.LstdFlags),
	}

	// Apply all options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Listen starts listening for JSON-RPC messages on the provided input and writes responses to the provided output.
// It runs until the context is cancelled or an error occurs.
// Returns an error if there are issues with reading input or writing output.
func (s *StdioServer) Listen(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	// Add in any custom context
	if s.contextFunc != nil {
		ctx = s.contextFunc(ctx)
	}

	reader := bufio.NewReader(stdin)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read a line from stdin
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					s.errLogger.Println("Input stream closed")
					return nil
				}
				s.errLogger.Printf("Error reading input: %v", err)
				return err
			}

			// Process the message in a separate goroutine
			// to avoid blocking the main loop
			go func(line string) {
				if err := s.processMessage(ctx, line, stdout); err != nil && err != io.EOF {
					s.errLogger.Printf("Error processing message: %v", err)

					// Try to send an error response
					errorResp := rest.JSONRPCResponse{
						JSONRPC: "2.0",
						ID:      nil, // We don't know the ID here
						Error: &rest.JSONRPCError{
							Code:    -32603,
							Message: fmt.Sprintf("Internal error: %v", err),
						},
					}
					if writeErr := s.writeResponse(errorResp, stdout); writeErr != nil {
						s.errLogger.Printf("Failed to write error response: %v", writeErr)
					}
				}
			}(line)
		}
	}
}

// processMessage handles a single JSON-RPC message and writes the response.
// It parses the message, processes it through the wrapped MCPServer, and writes any response.
// Returns an error if there are issues with message processing or response writing.
func (s *StdioServer) processMessage(ctx context.Context, line string, writer io.Writer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Trim whitespace from the line
	line = strings.TrimSpace(line)
	if line == "" {
		return nil // Skip empty lines
	}

	// Parse the message as raw JSON
	var rawMessage json.RawMessage
	if err := json.Unmarshal([]byte(line), &rawMessage); err != nil {
		response := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &rest.JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		}
		return s.writeResponse(response, writer)
	}

	// Handle the message using the wrapped server
	// We parse the message to determine the method and handle it
	var baseMessage struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      interface{} `json:"id"`
		Method  string      `json:"method"`
	}

	if err := json.Unmarshal(rawMessage, &baseMessage); err != nil {
		response := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &rest.JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		}
		return s.writeResponse(response, writer)
	}

	// Check if this is a notification (no ID field)
	// Notifications don't require responses
	if baseMessage.ID == nil && strings.HasPrefix(baseMessage.Method, "notifications/") {
		// Process notification but don't return a response
		// This is a notification message that doesn't expect a response
		s.errLogger.Printf("Received notification: %s", baseMessage.Method)
		return nil
	}

	// Create a dummy HTTP request and response for the server's HTTP handler
	w := httptest.NewRecorder()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "/jsonrpc", strings.NewReader(string(rawMessage)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r.Header.Set("Content-Type", "application/json")

	// Create a handler to process the request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			response := rest.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      baseMessage.ID,
				Error: &rest.JSONRPCError{
					Code:    -32700,
					Message: "Error reading request body",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Parse the request
		var req rest.JSONRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			resp := rest.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      baseMessage.ID,
				Error: &rest.JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Process the request based on method
		var result interface{}
		switch req.Method {
		case "initialize":
			name, version, instructions := s.server.GetServerInfo()
			result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
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

			if instructions != "" {
				result.(map[string]interface{})["instructions"] = instructions
			}

		case "ping":
			result = struct{}{}

		case "tools/list":
			// Handle tools listing
			tools, err := s.server.GetService().ListTools(r.Context())
			if err != nil {
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &rest.JSONRPCError{
						Code:    -32603,
						Message: fmt.Sprintf("Internal error: %v", err),
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
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

			result = map[string]interface{}{
				"tools": toolList,
			}

		case "tools/call":
			// Extract parameters
			params, ok := req.Params.(map[string]interface{})
			if !ok {
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &rest.JSONRPCError{
						Code:    -32602,
						Message: "Invalid params",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Get tool name
			toolName, ok := params["name"].(string)
			if !ok || toolName == "" {
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &rest.JSONRPCError{
						Code:    -32602,
						Message: "Missing or invalid 'name' parameter",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Get tool parameters - check both parameters and arguments fields
			toolParams, ok := params["parameters"].(map[string]interface{})
			if !ok {
				// Try arguments field if parameters is not available
				toolParams, ok = params["arguments"].(map[string]interface{})
				if !ok {
					toolParams = map[string]interface{}{}
				}
			}

			// Get available tools from the service
			tools, err := s.server.GetService().ListTools(r.Context())
			if err != nil {
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &rest.JSONRPCError{
						Code:    -32603,
						Message: fmt.Sprintf("Internal error: %v", err),
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Find the requested tool
			var foundTool *domain.Tool
			var toolFound bool

			for _, tool := range tools {
				if tool.Name == toolName {
					foundTool = tool
					toolFound = true
					break
				}
			}

			if !toolFound {
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &rest.JSONRPCError{
						Code:    -32601,
						Message: fmt.Sprintf("Tool not found: %s", toolName),
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Validate required parameters
			missingParams := []string{}
			for _, param := range foundTool.Parameters {
				if param.Required {
					paramValue, exists := toolParams[param.Name]
					if !exists || paramValue == nil {
						missingParams = append(missingParams, param.Name)
					}
				}
			}

			if len(missingParams) > 0 {
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &rest.JSONRPCError{
						Code:    -32602,
						Message: fmt.Sprintf("Missing required parameters: %s", strings.Join(missingParams, ", ")),
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// Handle all echo-related tools
			if strings.Contains(strings.ToLower(toolName), "echo") {
				// This is an echo tool
				message, ok := toolParams["message"].(string)
				if !ok || message == "" {
					resp := rest.JSONRPCResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error: &rest.JSONRPCError{
							Code:    -32602,
							Message: "Missing or invalid 'message' parameter",
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
					return
				}

				// Return the echoed message
				resp := rest.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": message,
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}

			// If we get here, the tool is not implemented
			s.errLogger.Printf("Tool '%s' is not implemented", toolName)
			resp := rest.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &rest.JSONRPCError{
					Code:    -32603,
					Message: fmt.Sprintf("Tool '%s' is registered but has no implementation", toolName),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		default:
			// Check if this is a notification method (starts with "notifications/")
			if strings.HasPrefix(req.Method, "notifications/") {
				// This is a notification that doesn't require a response
				s.errLogger.Printf("Processed notification: %s", req.Method)
				return
			}

			// Method not found
			resp := rest.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &rest.JSONRPCError{
					Code:    -32601,
					Message: fmt.Sprintf("Method '%s' not found", req.Method),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Send successful response if we get here
		resp := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Handle the request
	handler.ServeHTTP(w, r)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Fix potential JSON-RPC message format issues
	var responseObj map[string]interface{}
	if err := json.Unmarshal(respBody, &responseObj); err != nil {
		// Invalid JSON response
		s.errLogger.Printf("Error parsing response: %v", err)
		errorResp := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      baseMessage.ID, // Use the original ID
			Error: &rest.JSONRPCError{
				Code:    -32603,
				Message: "Internal error: invalid response format",
			},
		}
		respBody, _ = json.Marshal(errorResp)
	} else {
		// Ensure ID is properly set
		if id, exists := responseObj["id"]; exists && id == nil {
			if baseMessage.ID != nil {
				responseObj["id"] = baseMessage.ID
			} else {
				responseObj["id"] = ""
			}
			respBody, _ = json.Marshal(responseObj)
		}

		// Ensure there's either a result or an error
		_, hasResult := responseObj["result"]
		_, hasError := responseObj["error"]

		if !hasResult && !hasError {
			// Neither result nor error - add an empty result
			responseObj["result"] = map[string]interface{}{}
			respBody, _ = json.Marshal(responseObj)
		}
	}

	// Write the response
	if _, err = writer.Write(respBody); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	// Add a newline
	if _, err = writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("error writing newline: %w", err)
	}

	return nil
}

// writeResponse marshals and writes a JSON-RPC response message followed by a newline.
// Returns an error if marshaling or writing fails.
func (s *StdioServer) writeResponse(response interface{}, writer io.Writer) error {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Write response followed by newline
	if _, err := fmt.Fprintf(writer, "%s\n", responseBytes); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}

// ServeStdio is a convenience function that creates and starts a StdioServer with os.Stdin and os.Stdout.
// It sets up signal handling for graceful shutdown on SIGTERM and SIGINT.
// Returns an error if the server encounters any issues during operation.
func ServeStdio(server *rest.MCPServer, opts ...StdioOption) error {
	s := NewStdioServer(server, opts...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		s.errLogger.Printf("Received shutdown signal %v, stopping server...", sig)
		cancel()
	}()

	s.errLogger.Println("Starting MCP server in stdio mode")

	err := s.Listen(ctx, os.Stdin, os.Stdout)
	if err != nil && err != context.Canceled {
		s.errLogger.Printf("Server exited with error: %v", err)
		return err
	}

	s.errLogger.Println("Server shutdown complete")
	return nil
}
