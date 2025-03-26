package builder

import (
	"context"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

// ServerBuilder implements the Builder pattern for creating MCP servers
type ServerBuilder struct {
	name               string
	version            string
	instructions       string
	address            string
	resourceRepo       domain.ResourceRepository
	toolRepo           domain.ToolRepository
	promptRepo         domain.PromptRepository
	sessionRepo        domain.SessionRepository
	notificationSender domain.NotificationSender
}

// NewServerBuilder creates a new server builder with default values
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		name:         "MCP Server",
		version:      "1.0.0",
		instructions: "MCP Server for AI tools and resources",
		address:      ":8080",
		resourceRepo: server.NewInMemoryResourceRepository(),
		toolRepo:     server.NewInMemoryToolRepository(),
		promptRepo:   server.NewInMemoryPromptRepository(),
		sessionRepo:  server.NewInMemorySessionRepository(),
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

	return usecases.NewServerService(config)
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
	mcpServer := b.BuildMCPServer()
	return stdio.ServeStdio(mcpServer, opts...)
}
