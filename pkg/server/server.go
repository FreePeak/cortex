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
	"github.com/FreePeak/cortex/pkg/plugin"
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
// It supports both static tool registration and dynamic provider-based tools.
type MCPServer struct {
	name     string
	version  string
	tools    map[string]*types.Tool
	handlers map[string]ToolHandler
	registry plugin.Registry
	builder  *builder.ServerBuilder
	logger   *log.Logger
}

// NewMCPServer creates a new MCP server with the specified name and version.
func NewMCPServer(name, version string, logger *log.Logger) *MCPServer {
	if logger == nil {
		logger = log.Default()
	}

	registry := plugin.NewRegistry(logger)

	return &MCPServer{
		name:     name,
		version:  version,
		tools:    make(map[string]*types.Tool),
		handlers: make(map[string]ToolHandler),
		registry: registry,
		builder:  builder.NewServerBuilder().WithName(name).WithVersion(version),
		logger:   logger,
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

	// Store the original tool name to use for registration
	originalName := tool.Name

	// Log the tool names
	s.logger.Printf("Adding tool with name: %s ", originalName)

	s.tools[originalName] = tool
	s.handlers[originalName] = handler

	// Add tool to the internal builder with original name
	s.builder.AddTool(ctx, convertToInternalTool(tool))

	// Create an adapter to convert from our API to the internal API
	serviceAdapter := func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error) {
		// Log that the handler is being called
		s.logger.Printf("Service handler called for tool: %s", originalName)

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

	// Register with original name
	service.RegisterToolHandler(originalName, serviceAdapter)
	s.logger.Printf("Registered tool: %s", originalName)

	return nil
}

// RegisterProvider registers a tool provider with the server.
func (s *MCPServer) RegisterProvider(ctx context.Context, provider plugin.Provider) error {
	// Register the provider with the registry
	err := s.registry.RegisterProvider(ctx, provider)
	if err != nil {
		return fmt.Errorf("failed to register provider: %w", err)
	}

	// Get all tools from the provider and register them with the builder
	tools, err := provider.GetTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tools from provider: %w", err)
	}

	// Register each tool with the builder
	for _, tool := range tools {
		// Get original name
		originalName := tool.Name

		// Convert to internal tool
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
				Items:       param.Items,
			}
		}

		// Add the tool to the internal builder
		s.builder.AddTool(ctx, internalTool)

		// Create an adapter to convert from our API to the internal API
		serviceAdapter := func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error) {
			// Convert domain session to public session
			pubSession := &types.ClientSession{
				ID:        session.ID,
				UserAgent: session.UserAgent,
				Connected: session.Connected,
			}

			// Create request and execute the tool through the provider
			request := &plugin.ExecuteRequest{
				ToolName:   originalName,
				Parameters: params,
				Session:    pubSession,
			}

			// Find the provider for this tool
			_, provider, err := s.registry.GetTool(ctx, originalName)
			if err != nil {
				return nil, fmt.Errorf("failed to get tool %s: %w", originalName, err)
			}

			// Execute the tool through the provider
			response, err := provider.ExecuteTool(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("failed to execute tool %s: %w", originalName, err)
			}

			if response.Error != nil {
				return nil, response.Error
			}

			return response.Content, nil
		}

		// Get the service from the builder
		service := s.builder.BuildService()

		// Register with original name
		service.RegisterToolHandler(originalName, serviceAdapter)
		s.logger.Printf("Registered tool: %s", originalName)
	}

	return nil
}

// UnregisterProvider removes a tool provider from the server.
func (s *MCPServer) UnregisterProvider(ctx context.Context, providerID string) error {
	// Get the provider first to retrieve its tools
	provider, err := s.registry.GetProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider %s: %w", providerID, err)
	}

	// Get all tools from the provider
	tools, err := provider.GetTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tools from provider %s: %w", providerID, err)
	}

	// Since the internal service API doesn't expose a way to unregister tools directly,
	// we'll need to handle this differently. Let's just log it for now.
	for _, tool := range tools {
		s.logger.Printf("Note: Tool %s cannot be unregistered from existing service. A new service will be needed.", tool.Name)
	}

	// Unregister the provider from the registry
	err = s.registry.UnregisterProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to unregister provider %s: %w", providerID, err)
	}

	return nil
}

// ServeStdio serves the MCP server over standard I/O.
func (s *MCPServer) ServeStdio() error {
	// Check if logging is disabled
	disableLogging := os.Getenv("MCP_DISABLE_LOGGING") == "true" ||
		os.Getenv("DISABLE_LOGGING") == "true"

	// In STDIO mode, we must write all logs to stderr, as stdout is reserved for JSON-RPC messages
	if !disableLogging {
		// Only log to stderr if logging is enabled
		s.logger.Printf("Starting MCP server over stdio: %s v%s", s.name, s.version)
	}

	// Create stdio options
	var stdioOpts []stdio.StdioOption

	// Always use stderr for logging in STDIO mode
	stdioOpts = append(stdioOpts, stdio.WithErrorLogger(s.logger))

	// Log registered tools for debugging
	if !disableLogging {
		service := s.builder.BuildService()
		toolHandlers := service.GetAllToolHandlerNames()
		s.logger.Printf("Available tools in the server: %v", toolHandlers)
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
			Items:       param.Items,
		}
	}

	return internalTool
}
