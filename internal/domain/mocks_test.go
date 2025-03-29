package domain

import (
	"context"
	"sync"
)

// MockResourceRepository is a mock implementation of ResourceRepository
type MockResourceRepository struct {
	resources map[string]*Resource
	mu        sync.RWMutex
}

// NewMockResourceRepository creates a new MockResourceRepository
func NewMockResourceRepository() *MockResourceRepository {
	return &MockResourceRepository{
		resources: make(map[string]*Resource),
	}
}

// GetResource retrieves a resource by its URI
func (m *MockResourceRepository) GetResource(ctx context.Context, uri string) (*Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resource, ok := m.resources[uri]
	if !ok {
		return nil, NewResourceNotFoundError(uri)
	}
	return resource, nil
}

// ListResources returns all available resources
func (m *MockResourceRepository) ListResources(ctx context.Context) ([]*Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resources := make([]*Resource, 0, len(m.resources))
	for _, resource := range m.resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

// AddResource adds a new resource to the repository
func (m *MockResourceRepository) AddResource(ctx context.Context, resource *Resource) error {
	if resource == nil {
		return NewValidationError("resource", "cannot be nil")
	}
	if resource.URI == "" {
		return NewValidationError("uri", "cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.resources[resource.URI] = resource
	return nil
}

// DeleteResource removes a resource from the repository
func (m *MockResourceRepository) DeleteResource(ctx context.Context, uri string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.resources[uri]; !ok {
		return NewResourceNotFoundError(uri)
	}

	delete(m.resources, uri)
	return nil
}

// MockToolRepository is a mock implementation of ToolRepository
type MockToolRepository struct {
	tools map[string]*Tool
	mu    sync.RWMutex
}

// NewMockToolRepository creates a new MockToolRepository
func NewMockToolRepository() *MockToolRepository {
	return &MockToolRepository{
		tools: make(map[string]*Tool),
	}
}

// GetTool retrieves a tool by its name
func (m *MockToolRepository) GetTool(ctx context.Context, name string) (*Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, ok := m.tools[name]
	if !ok {
		return nil, NewToolNotFoundError(name)
	}
	return tool, nil
}

// ListTools returns all available tools
func (m *MockToolRepository) ListTools(ctx context.Context) ([]*Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]*Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools, nil
}

// AddTool adds a new tool to the repository
func (m *MockToolRepository) AddTool(ctx context.Context, tool *Tool) error {
	if tool == nil {
		return NewValidationError("tool", "cannot be nil")
	}
	if tool.Name == "" {
		return NewValidationError("name", "cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.tools[tool.Name] = tool
	return nil
}

// DeleteTool removes a tool from the repository
func (m *MockToolRepository) DeleteTool(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tools[name]; !ok {
		return NewToolNotFoundError(name)
	}

	delete(m.tools, name)
	return nil
}

// MockPromptRepository is a mock implementation of PromptRepository
type MockPromptRepository struct {
	prompts map[string]*Prompt
	mu      sync.RWMutex
}

// NewMockPromptRepository creates a new MockPromptRepository
func NewMockPromptRepository() *MockPromptRepository {
	return &MockPromptRepository{
		prompts: make(map[string]*Prompt),
	}
}

// GetPrompt retrieves a prompt by its name
func (m *MockPromptRepository) GetPrompt(ctx context.Context, name string) (*Prompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompt, ok := m.prompts[name]
	if !ok {
		return nil, NewPromptNotFoundError(name)
	}
	return prompt, nil
}

// ListPrompts returns all available prompts
func (m *MockPromptRepository) ListPrompts(ctx context.Context) ([]*Prompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompts := make([]*Prompt, 0, len(m.prompts))
	for _, prompt := range m.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

// AddPrompt adds a new prompt to the repository
func (m *MockPromptRepository) AddPrompt(ctx context.Context, prompt *Prompt) error {
	if prompt == nil {
		return NewValidationError("prompt", "cannot be nil")
	}
	if prompt.Name == "" {
		return NewValidationError("name", "cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.prompts[prompt.Name] = prompt
	return nil
}

// DeletePrompt removes a prompt from the repository
func (m *MockPromptRepository) DeletePrompt(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.prompts[name]; !ok {
		return NewPromptNotFoundError(name)
	}

	delete(m.prompts, name)
	return nil
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	sessions map[string]*ClientSession
	mu       sync.RWMutex
}

// NewMockSessionRepository creates a new MockSessionRepository
func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*ClientSession),
	}
}

// GetSession retrieves a session by its ID
func (m *MockSessionRepository) GetSession(ctx context.Context, id string) (*ClientSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil, NewSessionNotFoundError(id)
	}
	return session, nil
}

// ListSessions returns all active sessions
func (m *MockSessionRepository) ListSessions(ctx context.Context) ([]*ClientSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*ClientSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// AddSession adds a new session to the repository
func (m *MockSessionRepository) AddSession(ctx context.Context, session *ClientSession) error {
	if session == nil {
		return NewValidationError("session", "cannot be nil")
	}
	if session.ID == "" {
		return NewValidationError("id", "cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[session.ID] = session
	return nil
}

// DeleteSession removes a session from the repository
func (m *MockSessionRepository) DeleteSession(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[id]; !ok {
		return NewSessionNotFoundError(id)
	}

	delete(m.sessions, id)
	return nil
}

// MockNotificationSender is a mock implementation of NotificationSender
type MockNotificationSender struct {
	notifications     []*Notification
	sentNotifications map[string][]*Notification
	mu                sync.RWMutex
}

// NewMockNotificationSender creates a new MockNotificationSender
func NewMockNotificationSender() *MockNotificationSender {
	return &MockNotificationSender{
		notifications:     make([]*Notification, 0),
		sentNotifications: make(map[string][]*Notification),
	}
}

// SendNotification sends a notification to a specific client
func (m *MockNotificationSender) SendNotification(ctx context.Context, sessionID string, notification *Notification) error {
	if notification == nil {
		return NewValidationError("notification", "cannot be nil")
	}
	if sessionID == "" {
		return NewValidationError("sessionID", "cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sentNotifications[sessionID] == nil {
		m.sentNotifications[sessionID] = make([]*Notification, 0)
	}
	m.sentNotifications[sessionID] = append(m.sentNotifications[sessionID], notification)
	return nil
}

// BroadcastNotification sends a notification to all connected clients
func (m *MockNotificationSender) BroadcastNotification(ctx context.Context, notification *Notification) error {
	if notification == nil {
		return NewValidationError("notification", "cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.notifications = append(m.notifications, notification)
	return nil
}

// GetBroadcastNotifications returns all broadcast notifications
func (m *MockNotificationSender) GetBroadcastNotifications() []*Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.notifications
}

// GetSentNotifications returns all notifications sent to a specific session
func (m *MockNotificationSender) GetSentNotifications(sessionID string) []*Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.sentNotifications[sessionID]
}
