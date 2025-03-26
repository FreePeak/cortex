// Package domain defines the core business logic and entities for the MCP server.
package domain

import (
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
