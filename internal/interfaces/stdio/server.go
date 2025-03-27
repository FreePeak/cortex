// Package stdio provides the stdio interface for the MCP server.
package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
)

// Constants for JSON-RPC
const (
	JSONRPCVersion = "2.0"

	// Error codes
	ParseErrorCode     = -32700
	InvalidParamsCode  = -32602
	MethodNotFoundCode = -32601
	InternalErrorCode  = -32603
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
	processor   *MessageProcessor
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

	// Initialize the message processor
	s.processor = NewMessageProcessor(s.server, s.errLogger)

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

	// Process messages serially to avoid concurrent writes to stdout
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

			// Process message and get response
			response, processErr := s.processor.Process(ctx, line)

			// Handle processing errors
			if processErr != nil {
				if isTerminalError(processErr) {
					return processErr
				}

				s.errLogger.Printf("Error processing message: %v", processErr)

				// If we have a response (error response), send it
				if response != nil {
					if err := s.writeResponse(response, stdout); err != nil {
						s.errLogger.Printf("Error writing error response: %v", err)
						if isTerminalError(err) {
							return err
						}
					}
				}

				// Continue processing next messages for non-terminal errors
				continue
			}

			// Send successful response if we have one
			if response != nil {
				if err := s.writeResponse(response, stdout); err != nil {
					s.errLogger.Printf("Error writing response: %v", err)
					if isTerminalError(err) {
						return err
					}
				}
			}
		}
	}
}

// writeResponse marshals and writes a JSON-RPC response message followed by a newline.
// Returns an error if marshaling or writing fails.
func (s *StdioServer) writeResponse(response interface{}, writer io.Writer) error {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	// Write response
	n, err := writer.Write(responseBytes)
	if err != nil {
		return fmt.Errorf("error writing response (%d bytes): %w", n, err)
	}

	// Add a newline
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("error writing newline: %w", err)
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

// MessageProcessor handles JSON-RPC message processing
type MessageProcessor struct {
	server    *rest.MCPServer
	errLogger *log.Logger
	handlers  map[string]MethodHandler
}

// MethodHandler defines the interface for JSON-RPC method handlers
type MethodHandler interface {
	Handle(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError)
}

// MethodHandlerFunc is a function type that implements MethodHandler
type MethodHandlerFunc func(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError)

// Handle calls the handler function
func (f MethodHandlerFunc) Handle(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError) {
	return f(ctx, params, id)
}

// NewMessageProcessor creates a new message processor with registered handlers
func NewMessageProcessor(server *rest.MCPServer, logger *log.Logger) *MessageProcessor {
	p := &MessageProcessor{
		server:    server,
		errLogger: logger,
		handlers:  make(map[string]MethodHandler),
	}

	// Register standard handlers
	p.RegisterHandler("initialize", MethodHandlerFunc(p.handleInitialize))
	p.RegisterHandler("ping", MethodHandlerFunc(p.handlePing))
	p.RegisterHandler("tools/list", MethodHandlerFunc(p.handleToolsList))
	p.RegisterHandler("tools/call", MethodHandlerFunc(p.handleToolsCall))

	return p
}

// RegisterHandler registers a method handler
func (p *MessageProcessor) RegisterHandler(method string, handler MethodHandler) {
	p.handlers[method] = handler
}

// Process processes a JSON-RPC message and returns a response
func (p *MessageProcessor) Process(ctx context.Context, message string) (interface{}, error) {
	// Trim whitespace from the message
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, nil // Skip empty messages
	}

	// Create a timeout context for message processing
	msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse the message as a JSON-RPC request
	var baseMessage struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      interface{} `json:"id"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
	}

	if err := json.Unmarshal([]byte(message), &baseMessage); err != nil {
		return createErrorResponse(nil, ParseErrorCode, "Parse error"), nil
	}

	// Check if this is a notification (no ID field)
	// Notifications don't require responses
	if baseMessage.ID == nil && strings.HasPrefix(baseMessage.Method, "notifications/") {
		p.errLogger.Printf("Received notification: %s", baseMessage.Method)
		// Process notification but don't return a response
		return nil, nil
	}

	// Find handler for the method
	handler, exists := p.handlers[baseMessage.Method]

	// Handle notifications with a prefix
	if !exists && strings.HasPrefix(baseMessage.Method, "notifications/") {
		p.errLogger.Printf("Processed notification: %s", baseMessage.Method)
		return nil, nil
	}

	// Method not found
	if !exists {
		return createErrorResponse(
			baseMessage.ID,
			MethodNotFoundCode,
			fmt.Sprintf("Method '%s' not found", baseMessage.Method),
		), nil
	}

	// Execute the method handler
	result, jsonRpcErr := handler.Handle(msgCtx, baseMessage.Params, baseMessage.ID)
	if jsonRpcErr != nil {
		return createErrorResponseFromJSONRPCError(baseMessage.ID, jsonRpcErr), nil
	}

	// Create success response
	return createSuccessResponse(baseMessage.ID, result), nil
}

// Method handlers

func (p *MessageProcessor) handleInitialize(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError) {
	name, version, instructions := p.server.GetServerInfo()
	result := map[string]interface{}{
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
		result["instructions"] = instructions
	}

	return result, nil
}

func (p *MessageProcessor) handlePing(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError) {
	return struct{}{}, nil
}

func (p *MessageProcessor) handleToolsList(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError) {
	// Access the service through the server to get tools
	tools, err := p.server.GetService().ListTools(ctx)
	if err != nil {
		return nil, &domain.JSONRPCError{
			Code:    InternalErrorCode,
			Message: fmt.Sprintf("Internal error: %v", err),
		}
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

	return map[string]interface{}{
		"tools": toolList,
	}, nil
}

func (p *MessageProcessor) handleToolsCall(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError) {
	// Extract parameters
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, &domain.JSONRPCError{
			Code:    InvalidParamsCode,
			Message: "Invalid params",
		}
	}

	// Get tool name
	toolName, ok := paramsMap["name"].(string)
	if !ok || toolName == "" {
		return nil, &domain.JSONRPCError{
			Code:    InvalidParamsCode,
			Message: "Missing or invalid 'name' parameter",
		}
	}

	// Get tool parameters - check both parameters and arguments fields
	toolParams, ok := paramsMap["parameters"].(map[string]interface{})
	if !ok {
		// Try arguments field if parameters is not available
		toolParams, ok = paramsMap["arguments"].(map[string]interface{})
		if !ok {
			toolParams = map[string]interface{}{}
		}
	}

	// Get available tools from the service
	tools, err := p.server.GetService().ListTools(ctx)
	if err != nil {
		return nil, &domain.JSONRPCError{
			Code:    InternalErrorCode,
			Message: fmt.Sprintf("Internal error: %v", err),
		}
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
		return nil, &domain.JSONRPCError{
			Code:    MethodNotFoundCode,
			Message: fmt.Sprintf("Tool not found: %s", toolName),
		}
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
		return nil, &domain.JSONRPCError{
			Code:    InvalidParamsCode,
			Message: fmt.Sprintf("Missing required parameters: %s", strings.Join(missingParams, ", ")),
		}
	}

	// Handle different tool types using a strategy pattern
	var toolResult interface{}
	var toolErr error

	// Handle all echo-related tools
	if strings.Contains(strings.ToLower(toolName), "echo") {
		toolResult, toolErr = handleEchoTool(toolParams)
	} else {
		return nil, &domain.JSONRPCError{
			Code:    InternalErrorCode,
			Message: fmt.Sprintf("Tool '%s' is registered but has no implementation", toolName),
		}
	}

	if toolErr != nil {
		return nil, &domain.JSONRPCError{
			Code:    InternalErrorCode,
			Message: toolErr.Error(),
		}
	}

	return toolResult, nil
}

// Handle echo tool types
func handleEchoTool(params map[string]interface{}) (interface{}, error) {
	var message string

	// Extract message parameter
	messageVal, exists := params["message"]
	if !exists || messageVal == nil {
		return nil, fmt.Errorf("missing 'message' parameter")
	}

	// Convert to string based on type
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

	// Return formatted result using the MCP content format
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}, nil
}

// Helper functions for error handling and response creation

// isTerminalError determines if an error should cause the server to shut down
func isTerminalError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection closed") ||
		strings.Contains(errStr, "use of closed network connection")
}

// createSuccessResponse creates a standard JSON-RPC success response
func createSuccessResponse(id interface{}, result interface{}) map[string]interface{} {
	// Handle nil result case
	if result == nil {
		result = map[string]interface{}{}
	}

	return map[string]interface{}{
		"jsonrpc": JSONRPCVersion,
		"id":      id,
		"result":  result,
	}
}

// createErrorResponse creates a standard JSON-RPC error response
func createErrorResponse(id interface{}, code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": JSONRPCVersion,
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

// createErrorResponseFromJSONRPCError creates an error response from a JSONRPCError
func createErrorResponseFromJSONRPCError(id interface{}, err *domain.JSONRPCError) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": JSONRPCVersion,
		"id":      id,
		"error": map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
			"data":    err.Data,
		},
	}
}
