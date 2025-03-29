// Package builder provides the Builder pattern for creating MCP servers.
package builder

import (
	"context"

	internalBuilder "github.com/FreePeak/golang-mcp-server-sdk/internal/builder"
	internalDomain "github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
	"github.com/FreePeak/golang-mcp-server-sdk/pkg/types"
)

// ServerBuilder implements the Builder pattern for creating MCP servers.
type ServerBuilder struct {
	internal *internalBuilder.ServerBuilder
}

// NewServerBuilder creates a new server builder with default values.
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		internal: internalBuilder.NewServerBuilder(),
	}
}

// WithName sets the server name.
func (b *ServerBuilder) WithName(name string) *ServerBuilder {
	b.internal.WithName(name)
	return b
}

// WithVersion sets the server version.
func (b *ServerBuilder) WithVersion(version string) *ServerBuilder {
	b.internal.WithVersion(version)
	return b
}

// WithInstructions sets the server instructions.
func (b *ServerBuilder) WithInstructions(instructions string) *ServerBuilder {
	b.internal.WithInstructions(instructions)
	return b
}

// WithAddress sets the server address.
func (b *ServerBuilder) WithAddress(address string) *ServerBuilder {
	b.internal.WithAddress(address)
	return b
}

// WithResourceRepository sets the resource repository.
func (b *ServerBuilder) WithResourceRepository(repo types.ResourceRepository) *ServerBuilder {
	// Type adaptation from pkg to internal
	b.internal.WithResourceRepository(&resourceRepositoryAdapter{repo})
	return b
}

// WithToolRepository sets the tool repository.
func (b *ServerBuilder) WithToolRepository(repo types.ToolRepository) *ServerBuilder {
	// Type adaptation from pkg to internal
	b.internal.WithToolRepository(&toolRepositoryAdapter{repo})
	return b
}

// WithPromptRepository sets the prompt repository.
func (b *ServerBuilder) WithPromptRepository(repo types.PromptRepository) *ServerBuilder {
	// Type adaptation from pkg to internal
	b.internal.WithPromptRepository(&promptRepositoryAdapter{repo})
	return b
}

// WithSessionRepository sets the session repository.
func (b *ServerBuilder) WithSessionRepository(repo types.SessionRepository) *ServerBuilder {
	// Type adaptation from pkg to internal
	b.internal.WithSessionRepository(&sessionRepositoryAdapter{repo})
	return b
}

// WithNotificationSender sets the notification sender.
func (b *ServerBuilder) WithNotificationSender(sender types.NotificationSender) *ServerBuilder {
	// Type adaptation from pkg to internal
	b.internal.WithNotificationSender(&notificationSenderAdapter{sender})
	return b
}

// AddTool adds a tool to the server's tool repository.
func (b *ServerBuilder) AddTool(ctx context.Context, tool *types.Tool) *ServerBuilder {
	// Convert pkg type to internal type
	internalTool := &internalDomain.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  make([]internalDomain.ToolParameter, len(tool.Parameters)),
	}

	// Convert parameters
	for i, param := range tool.Parameters {
		internalTool.Parameters[i] = internalDomain.ToolParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	b.internal.AddTool(ctx, internalTool)
	return b
}

// AddResource adds a resource to the server's resource repository.
func (b *ServerBuilder) AddResource(ctx context.Context, resource *types.Resource) *ServerBuilder {
	// Convert pkg type to internal type
	internalResource := &internalDomain.Resource{
		URI:         resource.URI,
		Name:        resource.Name,
		Description: resource.Description,
		MIMEType:    resource.MIMEType,
	}

	b.internal.AddResource(ctx, internalResource)
	return b
}

// AddPrompt adds a prompt to the server's prompt repository.
func (b *ServerBuilder) AddPrompt(ctx context.Context, prompt *types.Prompt) *ServerBuilder {
	// Convert pkg type to internal type
	internalPrompt := &internalDomain.Prompt{
		Name:        prompt.Name,
		Description: prompt.Description,
		Template:    prompt.Template,
		Parameters:  make([]internalDomain.PromptParameter, len(prompt.Parameters)),
	}

	// Convert parameters
	for i, param := range prompt.Parameters {
		internalPrompt.Parameters[i] = internalDomain.PromptParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	b.internal.AddPrompt(ctx, internalPrompt)
	return b
}

// ServeStdio builds and starts serving a stdio server.
func (b *ServerBuilder) ServeStdio(opts ...stdio.StdioOption) error {
	return b.internal.ServeStdio(opts...)
}

// Adapter types to convert between pkg and internal types

// resourceRepositoryAdapter adapts a pkg ResourceRepository to an internal ResourceRepository.
type resourceRepositoryAdapter struct {
	repo types.ResourceRepository
}

func (a *resourceRepositoryAdapter) GetResource(ctx context.Context, uri string) (*internalDomain.Resource, error) {
	resource, err := a.repo.GetResource(ctx, uri)
	if err != nil {
		return nil, err
	}

	return &internalDomain.Resource{
		URI:         resource.URI,
		Name:        resource.Name,
		Description: resource.Description,
		MIMEType:    resource.MIMEType,
	}, nil
}

func (a *resourceRepositoryAdapter) ListResources(ctx context.Context) ([]*internalDomain.Resource, error) {
	resources, err := a.repo.ListResources(ctx)
	if err != nil {
		return nil, err
	}

	internalResources := make([]*internalDomain.Resource, len(resources))
	for i, resource := range resources {
		internalResources[i] = &internalDomain.Resource{
			URI:         resource.URI,
			Name:        resource.Name,
			Description: resource.Description,
			MIMEType:    resource.MIMEType,
		}
	}

	return internalResources, nil
}

func (a *resourceRepositoryAdapter) AddResource(ctx context.Context, resource *internalDomain.Resource) error {
	return a.repo.AddResource(ctx, &types.Resource{
		URI:         resource.URI,
		Name:        resource.Name,
		Description: resource.Description,
		MIMEType:    resource.MIMEType,
	})
}

func (a *resourceRepositoryAdapter) DeleteResource(ctx context.Context, uri string) error {
	return a.repo.DeleteResource(ctx, uri)
}

// toolRepositoryAdapter adapts a pkg ToolRepository to an internal ToolRepository.
type toolRepositoryAdapter struct {
	repo types.ToolRepository
}

func (a *toolRepositoryAdapter) GetTool(ctx context.Context, name string) (*internalDomain.Tool, error) {
	tool, err := a.repo.GetTool(ctx, name)
	if err != nil {
		return nil, err
	}

	internalTool := &internalDomain.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  make([]internalDomain.ToolParameter, len(tool.Parameters)),
	}

	for i, param := range tool.Parameters {
		internalTool.Parameters[i] = internalDomain.ToolParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	return internalTool, nil
}

func (a *toolRepositoryAdapter) ListTools(ctx context.Context) ([]*internalDomain.Tool, error) {
	tools, err := a.repo.ListTools(ctx)
	if err != nil {
		return nil, err
	}

	internalTools := make([]*internalDomain.Tool, len(tools))
	for i, tool := range tools {
		internalTools[i] = &internalDomain.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  make([]internalDomain.ToolParameter, len(tool.Parameters)),
		}

		for j, param := range tool.Parameters {
			internalTools[i].Parameters[j] = internalDomain.ToolParameter{
				Name:        param.Name,
				Description: param.Description,
				Type:        param.Type,
				Required:    param.Required,
			}
		}
	}

	return internalTools, nil
}

func (a *toolRepositoryAdapter) AddTool(ctx context.Context, tool *internalDomain.Tool) error {
	pkgTool := &types.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  make([]types.ToolParameter, len(tool.Parameters)),
	}

	for i, param := range tool.Parameters {
		pkgTool.Parameters[i] = types.ToolParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	return a.repo.AddTool(ctx, pkgTool)
}

func (a *toolRepositoryAdapter) DeleteTool(ctx context.Context, name string) error {
	return a.repo.DeleteTool(ctx, name)
}

// promptRepositoryAdapter adapts a pkg PromptRepository to an internal PromptRepository.
type promptRepositoryAdapter struct {
	repo types.PromptRepository
}

func (a *promptRepositoryAdapter) GetPrompt(ctx context.Context, name string) (*internalDomain.Prompt, error) {
	prompt, err := a.repo.GetPrompt(ctx, name)
	if err != nil {
		return nil, err
	}

	internalPrompt := &internalDomain.Prompt{
		Name:        prompt.Name,
		Description: prompt.Description,
		Template:    prompt.Template,
		Parameters:  make([]internalDomain.PromptParameter, len(prompt.Parameters)),
	}

	for i, param := range prompt.Parameters {
		internalPrompt.Parameters[i] = internalDomain.PromptParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	return internalPrompt, nil
}

func (a *promptRepositoryAdapter) ListPrompts(ctx context.Context) ([]*internalDomain.Prompt, error) {
	prompts, err := a.repo.ListPrompts(ctx)
	if err != nil {
		return nil, err
	}

	internalPrompts := make([]*internalDomain.Prompt, len(prompts))
	for i, prompt := range prompts {
		internalPrompts[i] = &internalDomain.Prompt{
			Name:        prompt.Name,
			Description: prompt.Description,
			Template:    prompt.Template,
			Parameters:  make([]internalDomain.PromptParameter, len(prompt.Parameters)),
		}

		for j, param := range prompt.Parameters {
			internalPrompts[i].Parameters[j] = internalDomain.PromptParameter{
				Name:        param.Name,
				Description: param.Description,
				Type:        param.Type,
				Required:    param.Required,
			}
		}
	}

	return internalPrompts, nil
}

func (a *promptRepositoryAdapter) AddPrompt(ctx context.Context, prompt *internalDomain.Prompt) error {
	pkgPrompt := &types.Prompt{
		Name:        prompt.Name,
		Description: prompt.Description,
		Template:    prompt.Template,
		Parameters:  make([]types.PromptParameter, len(prompt.Parameters)),
	}

	for i, param := range prompt.Parameters {
		pkgPrompt.Parameters[i] = types.PromptParameter{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			Required:    param.Required,
		}
	}

	return a.repo.AddPrompt(ctx, pkgPrompt)
}

func (a *promptRepositoryAdapter) DeletePrompt(ctx context.Context, name string) error {
	return a.repo.DeletePrompt(ctx, name)
}

// sessionRepositoryAdapter adapts a pkg SessionRepository to an internal SessionRepository.
type sessionRepositoryAdapter struct {
	repo types.SessionRepository
}

func (a *sessionRepositoryAdapter) GetSession(ctx context.Context, id string) (*internalDomain.ClientSession, error) {
	session, err := a.repo.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}

	return &internalDomain.ClientSession{
		ID:        session.ID,
		UserAgent: session.UserAgent,
		Connected: session.Connected,
	}, nil
}

func (a *sessionRepositoryAdapter) ListSessions(ctx context.Context) ([]*internalDomain.ClientSession, error) {
	sessions, err := a.repo.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	internalSessions := make([]*internalDomain.ClientSession, len(sessions))
	for i, session := range sessions {
		internalSessions[i] = &internalDomain.ClientSession{
			ID:        session.ID,
			UserAgent: session.UserAgent,
			Connected: session.Connected,
		}
	}

	return internalSessions, nil
}

func (a *sessionRepositoryAdapter) AddSession(ctx context.Context, session *internalDomain.ClientSession) error {
	return a.repo.AddSession(ctx, &types.ClientSession{
		ID:        session.ID,
		UserAgent: session.UserAgent,
		Connected: session.Connected,
	})
}

func (a *sessionRepositoryAdapter) DeleteSession(ctx context.Context, id string) error {
	return a.repo.DeleteSession(ctx, id)
}

// notificationSenderAdapter adapts a pkg NotificationSender to an internal NotificationSender.
type notificationSenderAdapter struct {
	sender types.NotificationSender
}

func (a *notificationSenderAdapter) SendNotification(ctx context.Context, sessionID string, notification *internalDomain.Notification) error {
	return a.sender.SendNotification(ctx, sessionID, &types.Notification{
		Method: notification.Method,
		Params: notification.Params,
	})
}

func (a *notificationSenderAdapter) BroadcastNotification(ctx context.Context, notification *internalDomain.Notification) error {
	return a.sender.BroadcastNotification(ctx, &types.Notification{
		Method: notification.Method,
		Params: notification.Params,
	})
}
