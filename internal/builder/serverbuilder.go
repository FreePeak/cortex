// Package builder provides a builder for the MCP server.
package builder

import (
	"context"

	"github.com/FreePeak/cortex/internal/domain"
	"github.com/FreePeak/cortex/internal/infrastructure/logging"
	"github.com/FreePeak/cortex/internal/infrastructure/server"
	"github.com/FreePeak/cortex/internal/interfaces/rest"
	"github.com/FreePeak/cortex/internal/interfaces/stdio"
	"github.com/FreePeak/cortex/internal/usecases"
)

// ServerBuilder is a builder for creating an MCP server.
type ServerBuilder struct {
	name               string
	version            string
	address            string
	instructions       string
	resourceRepo       domain.ResourceRepository
	toolRepo           domain.ToolRepository
	promptRepo         domain.PromptRepository
	sessionRepo        domain.SessionRepository
	notificationSender domain.NotificationSender

	// Maintain a single instance of the server service
	serverService *usecases.ServerService
}

// NewServerBuilder creates a new ServerBuilder.
func NewServerBuilder() *ServerBuilder {
	// Create default repositories
	toolRepo := server.NewInMemoryToolRepository()
	resourceRepo := server.NewInMemoryResourceRepository()
	promptRepo := server.NewInMemoryPromptRepository()
	sessionRepo := server.NewInMemorySessionRepository()

	return &ServerBuilder{
		name:          "MCP Server",
		version:       "1.0.0",
		address:       ":8080",
		instructions:  "MCP Server for AI tools and resources",
		resourceRepo:  resourceRepo,
		toolRepo:      toolRepo,
		promptRepo:    promptRepo,
		sessionRepo:   sessionRepo,
		serverService: nil, // Will be initialized when first needed
	}
}

// WithName sets the server name
func (b *ServerBuilder) WithName(name string) *ServerBuilder {
	b.name = name
	return b
}

// WithVersion sets the server version
func (b *ServerBuilder) WithVersion(version string) *ServerBuilder {
	b.version = version
	return b
}

// WithInstructions sets the server instructions
func (b *ServerBuilder) WithInstructions(instructions string) *ServerBuilder {
	b.instructions = instructions
	return b
}

// WithAddress sets the server address
func (b *ServerBuilder) WithAddress(address string) *ServerBuilder {
	b.address = address
	return b
}

// WithResourceRepository sets the resource repository
func (b *ServerBuilder) WithResourceRepository(repo domain.ResourceRepository) *ServerBuilder {
	b.resourceRepo = repo
	return b
}

// WithToolRepository sets the tool repository
func (b *ServerBuilder) WithToolRepository(repo domain.ToolRepository) *ServerBuilder {
	b.toolRepo = repo
	return b
}

// WithPromptRepository sets the prompt repository
func (b *ServerBuilder) WithPromptRepository(repo domain.PromptRepository) *ServerBuilder {
	b.promptRepo = repo
	return b
}

// WithSessionRepository sets the session repository
func (b *ServerBuilder) WithSessionRepository(repo domain.SessionRepository) *ServerBuilder {
	b.sessionRepo = repo
	return b
}

// WithNotificationSender sets the notification sender
func (b *ServerBuilder) WithNotificationSender(sender domain.NotificationSender) *ServerBuilder {
	b.notificationSender = sender
	return b
}

// AddTool adds a tool to the server's tool repository
func (b *ServerBuilder) AddTool(ctx context.Context, tool *domain.Tool) *ServerBuilder {
	if b.toolRepo != nil {
		_ = b.toolRepo.AddTool(ctx, tool)
	}
	return b
}

// AddResource adds a resource to the server's resource repository
func (b *ServerBuilder) AddResource(ctx context.Context, resource *domain.Resource) *ServerBuilder {
	if b.resourceRepo != nil {
		_ = b.resourceRepo.AddResource(ctx, resource)
	}
	return b
}

// AddPrompt adds a prompt to the server's prompt repository
func (b *ServerBuilder) AddPrompt(ctx context.Context, prompt *domain.Prompt) *ServerBuilder {
	if b.promptRepo != nil {
		_ = b.promptRepo.AddPrompt(ctx, prompt)
	}
	return b
}

// BuildService builds and returns the server service
func (b *ServerBuilder) BuildService() *usecases.ServerService {
	// If we already have a server service, return it
	if b.serverService != nil {
		return b.serverService
	}

	// Create notification sender if not provided
	if b.notificationSender == nil {
		b.notificationSender = server.NewNotificationSender("2.0")
	}

	// Create the server service config
	config := usecases.ServerConfig{
		Name:               b.name,
		Version:            b.version,
		Instructions:       b.instructions,
		ResourceRepo:       b.resourceRepo,
		ToolRepo:           b.toolRepo,
		PromptRepo:         b.promptRepo,
		SessionRepo:        b.sessionRepo,
		NotificationSender: b.notificationSender,
	}

	// Create and store the server service
	b.serverService = usecases.NewServerService(config)
	return b.serverService
}

// BuildMCPServer builds and returns an MCP server
func (b *ServerBuilder) BuildMCPServer() *rest.MCPServer {
	service := b.BuildService()
	return rest.NewMCPServer(service, b.address)
}

// BuildStdioServer builds a stdio server that uses the MCP server
func (b *ServerBuilder) BuildStdioServer(opts ...stdio.StdioOption) *stdio.StdioServer {
	mcpServer := b.BuildMCPServer()
	return stdio.NewStdioServer(mcpServer, opts...)
}

// ServeStdio builds and starts serving a stdio server
func (b *ServerBuilder) ServeStdio(opts ...stdio.StdioOption) error {
	// Create a default logger for stdio
	logger, err := logging.New(logging.Config{
		Level:       logging.InfoLevel,
		Development: true,
		OutputPaths: []string{"stderr"},
		InitialFields: logging.Fields{
			"component": "stdio-server",
		},
	})
	if err != nil {
		// If we can't create the logger, continue with the options provided
		// The stdio server will create its own default logger
	} else {
		// Prepend the logger option so it can be overridden by user-provided options
		opts = append([]stdio.StdioOption{stdio.WithLogger(logger)}, opts...)
	}

	mcpServer := b.BuildMCPServer()
	return stdio.ServeStdio(mcpServer, opts...)
}
