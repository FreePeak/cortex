package usecases

import (
	"context"
	"sync"
	"testing"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

// Test mocks

// MockResourceRepository is a mock implementation of domain.ResourceRepository
type MockResourceRepository struct {
	resources map[string]*domain.Resource
	mu        sync.RWMutex
}

// NewMockResourceRepository creates a new MockResourceRepository
func NewMockResourceRepository() *MockResourceRepository {
	return &MockResourceRepository{
		resources: make(map[string]*domain.Resource),
	}
}

// GetResource retrieves a resource by its URI
func (m *MockResourceRepository) GetResource(ctx context.Context, uri string) (*domain.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resource, ok := m.resources[uri]
	if !ok {
		return nil, domain.NewResourceNotFoundError(uri)
	}
	return resource, nil
}

// ListResources returns all available resources
func (m *MockResourceRepository) ListResources(ctx context.Context) ([]*domain.Resource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resources := make([]*domain.Resource, 0, len(m.resources))
	for _, resource := range m.resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

// AddResource adds a new resource to the repository
func (m *MockResourceRepository) AddResource(ctx context.Context, resource *domain.Resource) error {
	if resource == nil {
		return domain.NewValidationError("resource", "cannot be nil")
	}
	if resource.URI == "" {
		return domain.NewValidationError("uri", "cannot be empty")
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
		return domain.NewResourceNotFoundError(uri)
	}

	delete(m.resources, uri)
	return nil
}

// MockToolRepository is a mock implementation of domain.ToolRepository
type MockToolRepository struct {
	tools map[string]*domain.Tool
	mu    sync.RWMutex
}

// NewMockToolRepository creates a new MockToolRepository
func NewMockToolRepository() *MockToolRepository {
	return &MockToolRepository{
		tools: make(map[string]*domain.Tool),
	}
}

// GetTool retrieves a tool by its name
func (m *MockToolRepository) GetTool(ctx context.Context, name string) (*domain.Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, ok := m.tools[name]
	if !ok {
		return nil, domain.NewToolNotFoundError(name)
	}
	return tool, nil
}

// ListTools returns all available tools
func (m *MockToolRepository) ListTools(ctx context.Context) ([]*domain.Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]*domain.Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools, nil
}

// AddTool adds a new tool to the repository
func (m *MockToolRepository) AddTool(ctx context.Context, tool *domain.Tool) error {
	if tool == nil {
		return domain.NewValidationError("tool", "cannot be nil")
	}
	if tool.Name == "" {
		return domain.NewValidationError("name", "cannot be empty")
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
		return domain.NewToolNotFoundError(name)
	}

	delete(m.tools, name)
	return nil
}

// MockPromptRepository is a mock implementation of domain.PromptRepository
type MockPromptRepository struct {
	prompts map[string]*domain.Prompt
	mu      sync.RWMutex
}

// NewMockPromptRepository creates a new MockPromptRepository
func NewMockPromptRepository() *MockPromptRepository {
	return &MockPromptRepository{
		prompts: make(map[string]*domain.Prompt),
	}
}

// GetPrompt retrieves a prompt by its name
func (m *MockPromptRepository) GetPrompt(ctx context.Context, name string) (*domain.Prompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompt, ok := m.prompts[name]
	if !ok {
		return nil, domain.NewPromptNotFoundError(name)
	}
	return prompt, nil
}

// ListPrompts returns all available prompts
func (m *MockPromptRepository) ListPrompts(ctx context.Context) ([]*domain.Prompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompts := make([]*domain.Prompt, 0, len(m.prompts))
	for _, prompt := range m.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

// AddPrompt adds a new prompt to the repository
func (m *MockPromptRepository) AddPrompt(ctx context.Context, prompt *domain.Prompt) error {
	if prompt == nil {
		return domain.NewValidationError("prompt", "cannot be nil")
	}
	if prompt.Name == "" {
		return domain.NewValidationError("name", "cannot be empty")
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
		return domain.NewPromptNotFoundError(name)
	}

	delete(m.prompts, name)
	return nil
}

// MockSessionRepository is a mock implementation of domain.SessionRepository
type MockSessionRepository struct {
	sessions map[string]*domain.ClientSession
	mu       sync.RWMutex
}

// NewMockSessionRepository creates a new MockSessionRepository
func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*domain.ClientSession),
	}
}

// GetSession retrieves a session by its ID
func (m *MockSessionRepository) GetSession(ctx context.Context, id string) (*domain.ClientSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil, domain.NewSessionNotFoundError(id)
	}
	return session, nil
}

// ListSessions returns all active sessions
func (m *MockSessionRepository) ListSessions(ctx context.Context) ([]*domain.ClientSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*domain.ClientSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// AddSession adds a new session to the repository
func (m *MockSessionRepository) AddSession(ctx context.Context, session *domain.ClientSession) error {
	if session == nil {
		return domain.NewValidationError("session", "cannot be nil")
	}
	if session.ID == "" {
		return domain.NewValidationError("id", "cannot be empty")
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
		return domain.NewSessionNotFoundError(id)
	}

	delete(m.sessions, id)
	return nil
}

// MockNotificationSender is a mock implementation of domain.NotificationSender
type MockNotificationSender struct {
	notifications     []*domain.Notification
	sentNotifications map[string][]*domain.Notification
	mu                sync.RWMutex
}

// NewMockNotificationSender creates a new MockNotificationSender
func NewMockNotificationSender() *MockNotificationSender {
	return &MockNotificationSender{
		notifications:     make([]*domain.Notification, 0),
		sentNotifications: make(map[string][]*domain.Notification),
	}
}

// SendNotification sends a notification to a specific client
func (m *MockNotificationSender) SendNotification(ctx context.Context, sessionID string, notification *domain.Notification) error {
	if notification == nil {
		return domain.NewValidationError("notification", "cannot be nil")
	}
	if sessionID == "" {
		return domain.NewValidationError("sessionID", "cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sentNotifications[sessionID] == nil {
		m.sentNotifications[sessionID] = make([]*domain.Notification, 0)
	}
	m.sentNotifications[sessionID] = append(m.sentNotifications[sessionID], notification)
	return nil
}

// BroadcastNotification sends a notification to all connected clients
func (m *MockNotificationSender) BroadcastNotification(ctx context.Context, notification *domain.Notification) error {
	if notification == nil {
		return domain.NewValidationError("notification", "cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.notifications = append(m.notifications, notification)
	return nil
}

// GetBroadcastNotifications returns all broadcast notifications
func (m *MockNotificationSender) GetBroadcastNotifications() []*domain.Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.notifications
}

// GetSentNotifications returns all notifications sent to a specific session
func (m *MockNotificationSender) GetSentNotifications(sessionID string) []*domain.Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.sentNotifications[sessionID]
}

// Actual tests

func TestNewServerService(t *testing.T) {
	// Setup
	mockResourceRepo := NewMockResourceRepository()
	mockToolRepo := NewMockToolRepository()
	mockPromptRepo := NewMockPromptRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockNotificationSender := NewMockNotificationSender()

	// Test with valid config
	config := ServerConfig{
		Name:               "Test Server",
		Version:            "1.0.0",
		Instructions:       "Test instructions",
		ResourceRepo:       mockResourceRepo,
		ToolRepo:           mockToolRepo,
		PromptRepo:         mockPromptRepo,
		SessionRepo:        mockSessionRepo,
		NotificationSender: mockNotificationSender,
	}

	service := NewServerService(config)

	// Assert
	if service == nil {
		t.Fatal("NewServerService() returned nil")
	}

	name, version, instructions := service.ServerInfo()
	if name != config.Name {
		t.Errorf("ServerInfo() name = %v, want %v", name, config.Name)
	}
	if version != config.Version {
		t.Errorf("ServerInfo() version = %v, want %v", version, config.Version)
	}
	if instructions != config.Instructions {
		t.Errorf("ServerInfo() instructions = %v, want %v", instructions, config.Instructions)
	}
}

func TestServerService_Resource(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockResourceRepo := NewMockResourceRepository()
	service := createTestServerService(mockResourceRepo, nil, nil, nil, nil)

	// Test ListResources (empty)
	resources, err := service.ListResources(ctx)
	if err != nil {
		t.Errorf("ListResources() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("ListResources() returned %v resources, want 0", len(resources))
	}

	// Test AddResource
	resource := &domain.Resource{
		URI:         "test/resource",
		Name:        "Test Resource",
		Description: "A test resource",
		MIMEType:    "text/plain",
	}
	err = service.AddResource(ctx, resource)
	if err != nil {
		t.Errorf("AddResource() error = %v", err)
	}

	// Test GetResource
	retrieved, err := service.GetResource(ctx, resource.URI)
	if err != nil {
		t.Errorf("GetResource() error = %v", err)
	}
	if retrieved.URI != resource.URI {
		t.Errorf("GetResource().URI = %v, want %v", retrieved.URI, resource.URI)
	}
	if retrieved.Name != resource.Name {
		t.Errorf("GetResource().Name = %v, want %v", retrieved.Name, resource.Name)
	}

	// Test ListResources (with one resource)
	resources, err = service.ListResources(ctx)
	if err != nil {
		t.Errorf("ListResources() error = %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("ListResources() returned %v resources, want 1", len(resources))
	}

	// Test DeleteResource
	err = service.DeleteResource(ctx, resource.URI)
	if err != nil {
		t.Errorf("DeleteResource() error = %v", err)
	}

	// Verify resource is gone
	_, err = service.GetResource(ctx, resource.URI)
	if err == nil {
		t.Errorf("GetResource() should return error after deletion")
	}
}

func TestServerService_Tool(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockToolRepo := NewMockToolRepository()
	service := createTestServerService(nil, mockToolRepo, nil, nil, nil)

	// Test ListTools (empty)
	tools, err := service.ListTools(ctx)
	if err != nil {
		t.Errorf("ListTools() error = %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("ListTools() returned %v tools, want 0", len(tools))
	}

	// Test AddTool
	tool := &domain.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: []domain.ToolParameter{
			{
				Name:        "param1",
				Description: "A parameter",
				Type:        "string",
				Required:    true,
			},
		},
	}
	err = service.AddTool(ctx, tool)
	if err != nil {
		t.Errorf("AddTool() error = %v", err)
	}

	// Test GetTool
	retrieved, err := service.GetTool(ctx, tool.Name)
	if err != nil {
		t.Errorf("GetTool() error = %v", err)
	}
	if retrieved.Name != tool.Name {
		t.Errorf("GetTool().Name = %v, want %v", retrieved.Name, tool.Name)
	}
	if retrieved.Description != tool.Description {
		t.Errorf("GetTool().Description = %v, want %v", retrieved.Description, tool.Description)
	}

	// Test ListTools (with one tool)
	tools, err = service.ListTools(ctx)
	if err != nil {
		t.Errorf("ListTools() error = %v", err)
	}
	if len(tools) != 1 {
		t.Errorf("ListTools() returned %v tools, want 1", len(tools))
	}

	// Test DeleteTool
	err = service.DeleteTool(ctx, tool.Name)
	if err != nil {
		t.Errorf("DeleteTool() error = %v", err)
	}

	// Verify tool is gone
	_, err = service.GetTool(ctx, tool.Name)
	if err == nil {
		t.Errorf("GetTool() should return error after deletion")
	}
}

func TestServerService_Prompt(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockPromptRepo := NewMockPromptRepository()
	service := createTestServerService(nil, nil, mockPromptRepo, nil, nil)

	// Test ListPrompts (empty)
	prompts, err := service.ListPrompts(ctx)
	if err != nil {
		t.Errorf("ListPrompts() error = %v", err)
	}
	if len(prompts) != 0 {
		t.Errorf("ListPrompts() returned %v prompts, want 0", len(prompts))
	}

	// Test AddPrompt
	prompt := &domain.Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Template:    "Hello, {{name}}!",
		Parameters: []domain.PromptParameter{
			{
				Name:        "name",
				Description: "The name to greet",
				Type:        "string",
				Required:    true,
			},
		},
	}
	err = service.AddPrompt(ctx, prompt)
	if err != nil {
		t.Errorf("AddPrompt() error = %v", err)
	}

	// Test GetPrompt
	retrieved, err := service.GetPrompt(ctx, prompt.Name)
	if err != nil {
		t.Errorf("GetPrompt() error = %v", err)
	}
	if retrieved.Name != prompt.Name {
		t.Errorf("GetPrompt().Name = %v, want %v", retrieved.Name, prompt.Name)
	}
	if retrieved.Template != prompt.Template {
		t.Errorf("GetPrompt().Template = %v, want %v", retrieved.Template, prompt.Template)
	}

	// Test ListPrompts (with one prompt)
	prompts, err = service.ListPrompts(ctx)
	if err != nil {
		t.Errorf("ListPrompts() error = %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("ListPrompts() returned %v prompts, want 1", len(prompts))
	}

	// Test DeletePrompt
	err = service.DeletePrompt(ctx, prompt.Name)
	if err != nil {
		t.Errorf("DeletePrompt() error = %v", err)
	}

	// Verify prompt is gone
	_, err = service.GetPrompt(ctx, prompt.Name)
	if err == nil {
		t.Errorf("GetPrompt() should return error after deletion")
	}
}

func TestServerService_Session(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockSessionRepo := NewMockSessionRepository()
	service := createTestServerService(nil, nil, nil, mockSessionRepo, nil)

	// Create a test session
	session := domain.NewClientSession("test-user-agent")

	// Test RegisterSession
	err := service.RegisterSession(ctx, session)
	if err != nil {
		t.Errorf("RegisterSession() error = %v", err)
	}

	// Test UnregisterSession
	err = service.UnregisterSession(ctx, session.ID)
	if err != nil {
		t.Errorf("UnregisterSession() error = %v", err)
	}

	// Test UnregisterSession with non-existent ID
	err = service.UnregisterSession(ctx, "non-existent-id")
	if err == nil {
		t.Errorf("UnregisterSession() should return error for non-existent ID")
	}
}

func TestServerService_Notification(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockNotificationSender := NewMockNotificationSender()
	service := createTestServerService(nil, nil, nil, nil, mockNotificationSender)

	// Create a test session
	session := domain.NewClientSession("test-user-agent")

	// Create a test notification
	notification := &domain.Notification{
		Method: "test/notification",
		Params: map[string]interface{}{
			"key": "value",
		},
	}

	// Test SendNotification
	err := service.SendNotification(ctx, session.ID, notification)
	if err != nil {
		t.Errorf("SendNotification() error = %v", err)
	}

	// Test BroadcastNotification
	err = service.BroadcastNotification(ctx, notification)
	if err != nil {
		t.Errorf("BroadcastNotification() error = %v", err)
	}

	// Verify notifications were sent
	sentNotifications := mockNotificationSender.GetSentNotifications(session.ID)
	if len(sentNotifications) != 1 {
		t.Errorf("Expected 1 sent notification, got %d", len(sentNotifications))
	}

	broadcastNotifications := mockNotificationSender.GetBroadcastNotifications()
	if len(broadcastNotifications) != 1 {
		t.Errorf("Expected 1 broadcast notification, got %d", len(broadcastNotifications))
	}
}

func TestServerService_ResourceNotifications(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockResourceRepo := NewMockResourceRepository()
	mockNotificationSender := NewMockNotificationSender()
	service := createTestServerService(mockResourceRepo, nil, nil, nil, mockNotificationSender)

	// Test AddResource (should trigger notification)
	resource := &domain.Resource{
		URI:         "test/resource",
		Name:        "Test Resource",
		Description: "A test resource",
		MIMEType:    "text/plain",
	}
	err := service.AddResource(ctx, resource)
	if err != nil {
		t.Errorf("AddResource() error = %v", err)
	}

	// Verify notification was sent
	broadcastNotifications := mockNotificationSender.GetBroadcastNotifications()
	if len(broadcastNotifications) != 1 {
		t.Errorf("Expected 1 broadcast notification after AddResource, got %d", len(broadcastNotifications))
	}
	if broadcastNotifications[0].Method != "resources/list/changed" {
		t.Errorf("Expected notification method to be 'resources/list/changed', got %s", broadcastNotifications[0].Method)
	}

	// Test DeleteResource (should trigger notification)
	err = service.DeleteResource(ctx, resource.URI)
	if err != nil {
		t.Errorf("DeleteResource() error = %v", err)
	}

	// Verify second notification was sent
	broadcastNotifications = mockNotificationSender.GetBroadcastNotifications()
	if len(broadcastNotifications) != 2 {
		t.Errorf("Expected 2 broadcast notifications after DeleteResource, got %d", len(broadcastNotifications))
	}
	if broadcastNotifications[1].Method != "resources/list/changed" {
		t.Errorf("Expected notification method to be 'resources/list/changed', got %s", broadcastNotifications[1].Method)
	}
}

// Helper function to create a test server service
func createTestServerService(
	resourceRepo domain.ResourceRepository,
	toolRepo domain.ToolRepository,
	promptRepo domain.PromptRepository,
	sessionRepo domain.SessionRepository,
	notificationSender domain.NotificationSender,
) *ServerService {
	if resourceRepo == nil {
		resourceRepo = NewMockResourceRepository()
	}
	if toolRepo == nil {
		toolRepo = NewMockToolRepository()
	}
	if promptRepo == nil {
		promptRepo = NewMockPromptRepository()
	}
	if sessionRepo == nil {
		sessionRepo = NewMockSessionRepository()
	}
	if notificationSender == nil {
		notificationSender = NewMockNotificationSender()
	}

	config := ServerConfig{
		Name:               "Test Server",
		Version:            "1.0.0",
		Instructions:       "Test instructions",
		ResourceRepo:       resourceRepo,
		ToolRepo:           toolRepo,
		PromptRepo:         promptRepo,
		SessionRepo:        sessionRepo,
		NotificationSender: notificationSender,
	}

	return NewServerService(config)
}
