// Package types provides the core types for the MCP server SDK.
package types

import (
	"context"

	"github.com/google/uuid"
)

// ClientSession represents an active client connection to the MCP server.
type ClientSession struct {
	ID        string
	UserAgent string
	Connected bool
}

// NewClientSession creates a new ClientSession with a unique ID.
func NewClientSession(userAgent string) *ClientSession {
	return &ClientSession{
		ID:        uuid.New().String(),
		UserAgent: userAgent,
		Connected: true,
	}
}

// Resource represents a resource that can be requested by clients.
type Resource struct {
	URI         string
	Name        string
	Description string
	MIMEType    string
}

// ResourceContents represents the contents of a resource.
type ResourceContents struct {
	URI      string
	MIMEType string
	Content  []byte
	Text     string
}

// Tool represents a tool that can be called by clients.
type Tool struct {
	Name        string
	Description string
	Parameters  []ToolParameter
}

// ToolParameter defines a parameter for a tool.
type ToolParameter struct {
	Name        string
	Description string
	Type        string
	Required    bool
}

// ToolCall represents a request to execute a tool.
type ToolCall struct {
	Name       string
	Parameters map[string]interface{}
	Session    *ClientSession
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Data  interface{}
	Error error
}

// Prompt represents a prompt template that can be rendered.
type Prompt struct {
	Name        string
	Description string
	Template    string
	Parameters  []PromptParameter
}

// PromptParameter defines a parameter for a prompt template.
type PromptParameter struct {
	Name        string
	Description string
	Type        string
	Required    bool
}

// PromptRequest represents a request to render a prompt.
type PromptRequest struct {
	Name       string
	Parameters map[string]interface{}
	Session    *ClientSession
}

// PromptResult represents the result of a prompt rendering.
type PromptResult struct {
	Text  string
	Error error
}

// Notification represents a notification that can be sent to clients.
type Notification struct {
	Method string
	Params map[string]interface{}
}

// ResourceRepository defines the interface for managing resources.
type ResourceRepository interface {
	// GetResource retrieves a resource by its URI.
	GetResource(ctx context.Context, uri string) (*Resource, error)

	// ListResources returns all available resources.
	ListResources(ctx context.Context) ([]*Resource, error)

	// AddResource adds a new resource to the repository.
	AddResource(ctx context.Context, resource *Resource) error

	// DeleteResource removes a resource from the repository.
	DeleteResource(ctx context.Context, uri string) error
}

// ToolRepository defines the interface for managing tools.
type ToolRepository interface {
	// GetTool retrieves a tool by its name.
	GetTool(ctx context.Context, name string) (*Tool, error)

	// ListTools returns all available tools.
	ListTools(ctx context.Context) ([]*Tool, error)

	// AddTool adds a new tool to the repository.
	AddTool(ctx context.Context, tool *Tool) error

	// DeleteTool removes a tool from the repository.
	DeleteTool(ctx context.Context, name string) error
}

// PromptRepository defines the interface for managing prompts.
type PromptRepository interface {
	// GetPrompt retrieves a prompt by its name.
	GetPrompt(ctx context.Context, name string) (*Prompt, error)

	// ListPrompts returns all available prompts.
	ListPrompts(ctx context.Context) ([]*Prompt, error)

	// AddPrompt adds a new prompt to the repository.
	AddPrompt(ctx context.Context, prompt *Prompt) error

	// DeletePrompt removes a prompt from the repository.
	DeletePrompt(ctx context.Context, name string) error
}

// SessionRepository defines the interface for managing client sessions.
type SessionRepository interface {
	// GetSession retrieves a session by its ID.
	GetSession(ctx context.Context, id string) (*ClientSession, error)

	// ListSessions returns all active sessions.
	ListSessions(ctx context.Context) ([]*ClientSession, error)

	// AddSession adds a new session to the repository.
	AddSession(ctx context.Context, session *ClientSession) error

	// DeleteSession removes a session from the repository.
	DeleteSession(ctx context.Context, id string) error
}

// NotificationSender defines the interface for sending notifications to clients.
type NotificationSender interface {
	// SendNotification sends a notification to a specific client.
	SendNotification(ctx context.Context, sessionID string, notification *Notification) error

	// BroadcastNotification sends a notification to all connected clients.
	BroadcastNotification(ctx context.Context, notification *Notification) error
}
