// Package usecases implements the application business logic for the MCP server.
package usecases

import (
	"context"

	"github.com/FreePeak/cortex/internal/domain"
)

// ToolHandlerFunc defines a function type for handling tool calls
type ToolHandlerFunc func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error)

// ServerService handles business logic for the MCP server.
type ServerService struct {
	name               string
	version            string
	instructions       string
	resourceRepo       domain.ResourceRepository
	toolRepo           domain.ToolRepository
	promptRepo         domain.PromptRepository
	sessionRepo        domain.SessionRepository
	notificationSender domain.NotificationSender
	toolHandlers       map[string]ToolHandlerFunc // Map of tool names to handler functions
}

// ServerConfig contains configuration for the ServerService.
type ServerConfig struct {
	Name               string
	Version            string
	Instructions       string
	ResourceRepo       domain.ResourceRepository
	ToolRepo           domain.ToolRepository
	PromptRepo         domain.PromptRepository
	SessionRepo        domain.SessionRepository
	NotificationSender domain.NotificationSender
}

// NewServerService creates a new ServerService with the given repositories and configuration.
func NewServerService(config ServerConfig) *ServerService {
	service := &ServerService{
		name:               config.Name,
		version:            config.Version,
		instructions:       config.Instructions,
		resourceRepo:       config.ResourceRepo,
		toolRepo:           config.ToolRepo,
		promptRepo:         config.PromptRepo,
		sessionRepo:        config.SessionRepo,
		notificationSender: config.NotificationSender,
		toolHandlers:       make(map[string]ToolHandlerFunc),
	}

	// No longer automatically register built-in tool handlers
	// Clients must register any tools they need

	return service
}

// RegisterToolHandler registers a handler for a specific tool
func (s *ServerService) RegisterToolHandler(name string, handler ToolHandlerFunc) {
	// Register with original name
	s.toolHandlers[name] = handler

	// Register with prefixed name if it doesn't already start with the prefix
	prefixedName := "cortex_" + name
	// Only register if there's not already a handler with this name
	if _, exists := s.toolHandlers[prefixedName]; !exists {
		s.toolHandlers[prefixedName] = handler
	}
}

// GetToolHandler retrieves a handler for a specific tool
func (s *ServerService) GetToolHandler(name string) ToolHandlerFunc {
	// Try to get the handler with the exact name
	if handler, exists := s.toolHandlers[name]; exists {
		return handler
	}

	// If the name starts with "cortex_", try without the prefix
	if len(name) > 7 && name[:7] == "cortex_" {
		unprefixedName := name[7:]
		if handler, exists := s.toolHandlers[unprefixedName]; exists {
			return handler
		}
	}

	// If the name doesn't have the prefix, try with the prefix
	prefixedName := "cortex_" + name
	return s.toolHandlers[prefixedName]
}

// GetAllToolHandlerNames returns a slice of all registered tool handler names
func (s *ServerService) GetAllToolHandlerNames() []string {
	names := make([]string, 0, len(s.toolHandlers))
	for name := range s.toolHandlers {
		names = append(names, name)
	}
	return names
}

// ServerInfo returns information about the server.
func (s *ServerService) ServerInfo() (string, string, string) {
	return s.name, s.version, s.instructions
}

// ListResources returns all available resources.
func (s *ServerService) ListResources(ctx context.Context) ([]*domain.Resource, error) {
	return s.resourceRepo.ListResources(ctx)
}

// GetResource returns a resource by its URI.
func (s *ServerService) GetResource(ctx context.Context, uri string) (*domain.Resource, error) {
	return s.resourceRepo.GetResource(ctx, uri)
}

// AddResource adds a new resource.
func (s *ServerService) AddResource(ctx context.Context, resource *domain.Resource) error {
	// Notify clients about resource list change after adding
	defer s.notifyResourceListChanged(ctx)
	return s.resourceRepo.AddResource(ctx, resource)
}

// DeleteResource removes a resource.
func (s *ServerService) DeleteResource(ctx context.Context, uri string) error {
	// Notify clients about resource list change after deletion
	defer s.notifyResourceListChanged(ctx)
	return s.resourceRepo.DeleteResource(ctx, uri)
}

// ListTools returns all available tools.
func (s *ServerService) ListTools(ctx context.Context) ([]*domain.Tool, error) {
	return s.toolRepo.ListTools(ctx)
}

// GetTool returns a tool by its name.
func (s *ServerService) GetTool(ctx context.Context, name string) (*domain.Tool, error) {
	// Try to get the tool with the provided name
	tool, err := s.toolRepo.GetTool(ctx, name)
	if err == nil {
		return tool, nil
	}

	// If the name doesn't have the prefix, try with the prefix
	prefixedName := "cortex_" + name
	tool, err = s.toolRepo.GetTool(ctx, prefixedName)
	if err == nil {
		return tool, nil
	}

	// If all attempts fail, return the original error
	return nil, err
}

// AddTool adds a new tool.
func (s *ServerService) AddTool(ctx context.Context, tool *domain.Tool) error {
	// Notify clients about tool list change after adding
	defer s.notifyToolListChanged(ctx)
	return s.toolRepo.AddTool(ctx, tool)
}

// DeleteTool removes a tool.
func (s *ServerService) DeleteTool(ctx context.Context, name string) error {
	// Notify clients about tool list change after deletion
	defer s.notifyToolListChanged(ctx)
	return s.toolRepo.DeleteTool(ctx, name)
}

// ListPrompts returns all available prompts.
func (s *ServerService) ListPrompts(ctx context.Context) ([]*domain.Prompt, error) {
	return s.promptRepo.ListPrompts(ctx)
}

// GetPrompt returns a prompt by its name.
func (s *ServerService) GetPrompt(ctx context.Context, name string) (*domain.Prompt, error) {
	return s.promptRepo.GetPrompt(ctx, name)
}

// AddPrompt adds a new prompt.
func (s *ServerService) AddPrompt(ctx context.Context, prompt *domain.Prompt) error {
	// Notify clients about prompt list change after adding
	defer s.notifyPromptListChanged(ctx)
	return s.promptRepo.AddPrompt(ctx, prompt)
}

// DeletePrompt removes a prompt.
func (s *ServerService) DeletePrompt(ctx context.Context, name string) error {
	// Notify clients about prompt list change after deletion
	defer s.notifyPromptListChanged(ctx)
	return s.promptRepo.DeletePrompt(ctx, name)
}

// RegisterSession adds a new client session.
func (s *ServerService) RegisterSession(ctx context.Context, session *domain.ClientSession) error {
	return s.sessionRepo.AddSession(ctx, session)
}

// UnregisterSession removes a client session.
func (s *ServerService) UnregisterSession(ctx context.Context, id string) error {
	return s.sessionRepo.DeleteSession(ctx, id)
}

// SendNotification sends a notification to a specific client.
func (s *ServerService) SendNotification(ctx context.Context, sessionID string, notification *domain.Notification) error {
	return s.notificationSender.SendNotification(ctx, sessionID, notification)
}

// BroadcastNotification sends a notification to all connected clients.
func (s *ServerService) BroadcastNotification(ctx context.Context, notification *domain.Notification) error {
	return s.notificationSender.BroadcastNotification(ctx, notification)
}

// Helper methods for sending specific notifications

func (s *ServerService) notifyResourceListChanged(ctx context.Context) {
	notification := &domain.Notification{
		Method: "resources/list/changed",
		Params: map[string]interface{}{},
	}
	_ = s.BroadcastNotification(ctx, notification)
}

func (s *ServerService) notifyToolListChanged(ctx context.Context) {
	notification := &domain.Notification{
		Method: "tools/list/changed",
		Params: map[string]interface{}{},
	}
	_ = s.BroadcastNotification(ctx, notification)
}

func (s *ServerService) notifyPromptListChanged(ctx context.Context) {
	notification := &domain.Notification{
		Method: "prompts/list/changed",
		Params: map[string]interface{}{},
	}
	_ = s.BroadcastNotification(ctx, notification)
}
