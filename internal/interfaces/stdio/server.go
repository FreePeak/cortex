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

	"github.com/FreePeak/cortex/internal/domain"
	"github.com/FreePeak/cortex/internal/infrastructure/logging"
	"github.com/FreePeak/cortex/internal/interfaces/rest"
	"github.com/google/uuid"
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
	logger      *logging.Logger
	contextFunc StdioContextFunc
	processor   *MessageProcessor
}

// StdioOption defines a function type for configuring StdioServer
type StdioOption func(*StdioServer)

// WithLogger sets the logger for the server
func WithLogger(logger *logging.Logger) StdioOption {
	return func(s *StdioServer) {
		s.logger = logger
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

// WithErrorLogger is kept for backwards compatibility
// It will create a custom logger that wraps the standard log.Logger
func WithErrorLogger(stdLogger *log.Logger) StdioOption {
	return func(s *StdioServer) {
		// Create a development logger that always outputs to stderr
		// In STDIO mode, stdout is reserved for JSON-RPC messages only
		logger, err := logging.New(logging.Config{
			Level:       logging.InfoLevel,
			Development: true,
			OutputPaths: []string{"stderr"}, // Force stderr for all logging
		})
		if err != nil {
			// If we can't create the logger, use the default one
			return
		}
		s.logger = logger
	}
}

// NewStdioServer creates a new stdio server wrapper around an MCPServer.
// It initializes the server with a default logger that logs to stderr.
func NewStdioServer(server *rest.MCPServer, opts ...StdioOption) *StdioServer {
	// Create default logger - always use stderr for STDIO servers
	defaultLogger, err := logging.New(logging.Config{
		Level:       logging.InfoLevel,
		Development: true,
		OutputPaths: []string{"stderr"}, // Force stderr for all logging output
		InitialFields: logging.Fields{
			"component": "stdio-server",
		},
	})
	if err != nil {
		// Fallback to a simple default logger if we can't create the structured one
		defaultLogger = logging.Default()
	}

	s := &StdioServer{
		server: server,
		logger: defaultLogger,
	}

	// Apply all options
	for _, opt := range opts {
		opt(s)
	}

	// Initialize the message processor
	s.processor = NewMessageProcessor(s.server, s.logger)

	return s
}

// Listen starts listening for JSON-RPC messages on the provided input and writes responses to the provided output.
// It runs until the context is canceled or an error occurs.
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
					s.logger.Info("Input stream closed")
					return nil
				}
				s.logger.Error("Error reading input", logging.Fields{"error": err})
				return err
			}

			// Process message and get response
			response, processErr := s.processor.Process(ctx, line)

			// Handle processing errors
			if processErr != nil {
				if isTerminalError(processErr) {
					return processErr
				}

				s.logger.Error("Error processing message", logging.Fields{"error": processErr})

				// If we have a response (error response), send it
				if response != nil {
					if err := s.writeResponse(response, stdout); err != nil {
						s.logger.Error("Error writing error response", logging.Fields{"error": err})
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
					s.logger.Error("Error writing response", logging.Fields{"error": err})
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
		s.logger.Info("Received shutdown signal, stopping server...", logging.Fields{"signal": sig.String()})
		cancel()
	}()

	s.logger.Info("Starting MCP server in stdio mode")

	err := s.Listen(ctx, os.Stdin, os.Stdout)
	if err != nil && err != context.Canceled {
		s.logger.Error("Server exited with error", logging.Fields{"error": err})
		return err
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

// MessageProcessor handles JSON-RPC message processing
type MessageProcessor struct {
	server   *rest.MCPServer
	logger   *logging.Logger
	handlers map[string]MethodHandler
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
func NewMessageProcessor(server *rest.MCPServer, logger *logging.Logger) *MessageProcessor {
	p := &MessageProcessor{
		server:   server,
		logger:   logger,
		handlers: make(map[string]MethodHandler),
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
		p.logger.Info("Received notification", logging.Fields{"method": baseMessage.Method})
		// Process notification but don't return a response
		return nil, nil
	}

	// Find handler for the method
	handler, exists := p.handlers[baseMessage.Method]

	// Handle notifications with a prefix
	if !exists && strings.HasPrefix(baseMessage.Method, "notifications/") {
		p.logger.Info("Processed notification", logging.Fields{"method": baseMessage.Method})
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

	// Create a client session for the tool handler
	clientSession := &domain.ClientSession{
		ID:        generateSessionID(),
		Connected: true,
	}

	// Access the service to get the tool handler
	service := p.server.GetService()

	// Try to get a registered handler for this tool
	handler := service.GetToolHandler(toolName)
	if handler != nil {
		// We have a registered handler, use it
		p.logger.Info("Using registered handler for tool", logging.Fields{"tool": toolName})
		result, err := handler(ctx, toolParams, clientSession)
		if err != nil {
			p.logger.Error("Error executing tool handler", logging.Fields{"tool": toolName, "error": err})
			return nil, &domain.JSONRPCError{
				Code:    InternalErrorCode,
				Message: fmt.Sprintf("Error executing tool: %v", err),
			}
		}

		p.logger.Info("Tool executed successfully", logging.Fields{"tool": toolName})
		return result, nil
	}

	// No handler found - log available handlers and return error
	availableHandlers := service.GetAllToolHandlerNames()
	p.logger.Warn("No registered handler found for tool", logging.Fields{
		"tool":              toolName,
		"availableHandlers": fmt.Sprintf("%+v", availableHandlers),
	})

	return nil, &domain.JSONRPCError{
		Code:    InternalErrorCode,
		Message: fmt.Sprintf("Tool '%s' is registered but has no implementation", toolName),
	}
}

// generateSessionID creates a unique session ID
func generateSessionID() string {
	return uuid.New().String()
}

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
