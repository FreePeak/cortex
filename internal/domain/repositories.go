package domain

import "context"

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
