// Package server provides the MCP server implementation.
package server

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FreePeak/cortex/internal/builder"
	"github.com/FreePeak/cortex/internal/domain"
	"github.com/FreePeak/cortex/internal/interfaces/stdio"
	"github.com/FreePeak/cortex/pkg/types"
)

// ToolHandler is a function that handles tool calls.
type ToolHandler func(ctx context.Context, request ToolCallRequest) (interface{}, error)

// ToolCallRequest represents a request to execute a tool.
type ToolCallRequest struct {
	Name       string
	Parameters map[string]interface{}
	Session    *types.ClientSession
}

// MCPServer represents an MCP server that can be used to handle MCP protocol messages.
type MCPServer struct {
	name     string
	version  string
	tools    map[string]*types.Tool
	handlers map[string]ToolHandler
	builder  *builder.ServerBuilder
}

// NewMCPServer creates a new MCP server with the specified name and version.
func NewMCPServer(name, version string) *MCPServer {
	return &MCPServer{
		name:     name,
		version:  version,
		tools:    make(map[string]*types.Tool),
		handlers: make(map[string]ToolHandler),
		builder:  builder.NewServerBuilder().WithName(name).WithVersion(version),
	}
}

// AddTool adds a tool to the MCP server.
func (s *MCPServer) AddTool(ctx context.Context, tool *types.Tool, handler ToolHandler) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Log the original tool name
	log.Printf("Adding tool with original name: %s", tool.Name)

	// Store the original tool name to use for registration
	originalName := tool.Name

	// Store the tool and its handler using the original name
	s.tools[originalName] = tool
	s.handlers[originalName] = handler

	// Add tool to the internal builder
	s.builder.AddTool(ctx, convertToInternalTool(tool))

	// Register the tool handler with the ServerService to make it available to the HTTP/SSE server
	// Create an adapter to convert from our API to the internal API
	serviceAdapter := func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error) {
		// Log that the handler is being called
		log.Printf("Service handler called for tool: %s", originalName)

		// Convert domain session to public session
		pubSession := &types.ClientSession{
			ID:        session.ID,
			UserAgent: session.UserAgent,
			Connected: session.Connected,
		}

		// Create request and call the handler
		request := ToolCallRequest{
			Name:       originalName,
			Parameters: params,
			Session:    pubSession,
		}

		return handler(ctx, request)
	}

	// Get the service from the builder
	service := s.builder.BuildService()
	service.RegisterToolHandler(originalName, serviceAdapter)

	log.Printf("Registered tool: %s", originalName)
	return nil
}

// RegisterToolHandler registers a handler for the specified tool.
func (s *MCPServer) RegisterToolHandler(name string, handler ToolHandler) error {
	if _, exists := s.tools[name]; !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	s.handlers[name] = handler

	// Get the service from the builder
	service := s.builder.BuildService()

	// Create an adapter to convert from our API to the internal API
	serviceAdapter := func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error) {
		// Convert domain session to public session
		pubSession := &types.ClientSession{
			ID:        session.ID,
			UserAgent: session.UserAgent,
			Connected: session.Connected,
		}

		// Create request and call the handler
		request := ToolCallRequest{
			Name:       name,
			Parameters: params,
			Session:    pubSession,
		}

		return handler(ctx, request)
	}

	service.RegisterToolHandler(name, serviceAdapter)

	return nil
}

// ServeStdio serves the MCP server over standard I/O.
func (s *MCPServer) ServeStdio() error {
	debugMode := os.Getenv("CORTEX_DEBUG") == "1"

	if debugMode {
		log.Printf("ServeStdio called with %d registered tools", len(s.handlers))
	}

	// Create stdio options with tool handlers
	var stdioOpts []stdio.StdioOption

	// Add the default error logger
	stdioOpts = append(stdioOpts, stdio.WithErrorLogger(log.Default()))

	// First, create a map of toolHandlers
	toolHandlersMap := make(map[string]func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error))

	// Add all tool handlers to the map
	for name, handler := range s.handlers {
		// Only register tools with "cortex_" prefix or manually update the name to include prefix
		toolName := name
		if len(toolName) < 7 || toolName[:7] != "cortex_" {
			if debugMode {
				log.Printf("Skipping non-prefixed tool: %s", toolName)
			}
			// Skip non-prefixed tools
			continue
		}

		toolHandler := handler

		if debugMode {
			log.Printf("Registering tool handler for %s", toolName)
		}

		// Create an adapter function that converts between our API and the internal API
		adapter := func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error) {
			// Convert domain session to public session
			pubSession := &types.ClientSession{
				ID:        session.ID,
				UserAgent: session.UserAgent,
				Connected: session.Connected,
			}

			if debugMode {
				log.Printf("Tool handler called for %s with params %v", toolName, params)
			}

			// Create request and call the handler
			request := ToolCallRequest{
				Name:       toolName,
				Parameters: params,
				Session:    pubSession,
			}

			return toolHandler(ctx, request)
		}

		// Add the tool handler to the map
		toolHandlersMap[toolName] = adapter

		// Add the tool handler as an option
		stdioOpts = append(stdioOpts, stdio.WithToolHandler(toolName, adapter))
	}

	// Add an option to directly set all tool handlers at once
	if len(toolHandlersMap) > 0 {
		stdioOpts = append(stdioOpts, stdio.WithAllToolHandlers(toolHandlersMap))
		if debugMode {
			log.Printf("Registered %d tool handlers with WithAllToolHandlers", len(toolHandlersMap))
		}
	}

	// Start the stdio server with our custom handler
	return s.builder.ServeStdio(stdioOpts...)
}

// SetAddress sets the HTTP address for the server.
func (s *MCPServer) SetAddress(addr string) {
	s.builder.WithAddress(addr)
}

// GetAddress returns the HTTP address for the server.
func (s *MCPServer) GetAddress() string {
	// Build the MCP server to get the address
	restServer := s.builder.BuildMCPServer()
	return restServer.GetAddress()
}

// ServeHTTP starts the HTTP server.
func (s *MCPServer) ServeHTTP() error {
	// Create an HTTP server with all our tools already registered through the builder
	mcpServer := s.builder.BuildMCPServer()

	// The tools are already registered with the server through the builder pattern
	// when we called AddTool on our MCPServer, which called AddTool on the builder
	// Additionally, we've already registered the tool handlers with the ServerService

	// Start the HTTP server
	return mcpServer.Start()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *MCPServer) Shutdown(ctx context.Context) error {
	// Build the MCP server to get access to the Stop method
	mcpServer := s.builder.BuildMCPServer()
	return mcpServer.Stop(ctx)
}

// Helper function to convert a public tool to an internal tool
func convertToInternalTool(tool *types.Tool) *domain.Tool {
	internalTool := &domain.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  make([]domain.ToolParameter, len(tool.Parameters)),
	}

	for i, param := range tool.Parameters {
		internalTool.Parameters[i] = domain.ToolParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	return internalTool
}
