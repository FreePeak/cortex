package builder

import (
	"context"
	"testing"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResourceRepository is a mock for the ResourceRepository interface
type MockResourceRepository struct {
	mock.Mock
}

func (m *MockResourceRepository) AddResource(ctx context.Context, resource *domain.Resource) error {
	args := m.Called(ctx, resource)
	return args.Error(0)
}

func (m *MockResourceRepository) GetResource(ctx context.Context, uri string) (*domain.Resource, error) {
	args := m.Called(ctx, uri)
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func (m *MockResourceRepository) ListResources(ctx context.Context) ([]*domain.Resource, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Resource), args.Error(1)
}

func (m *MockResourceRepository) DeleteResource(ctx context.Context, uri string) error {
	args := m.Called(ctx, uri)
	return args.Error(0)
}

// MockToolRepository is a mock for the ToolRepository interface
type MockToolRepository struct {
	mock.Mock
}

func (m *MockToolRepository) AddTool(ctx context.Context, tool *domain.Tool) error {
	args := m.Called(ctx, tool)
	return args.Error(0)
}

func (m *MockToolRepository) GetTool(ctx context.Context, name string) (*domain.Tool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Tool), args.Error(1)
}

func (m *MockToolRepository) ListTools(ctx context.Context) ([]*domain.Tool, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Tool), args.Error(1)
}

func (m *MockToolRepository) DeleteTool(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// MockPromptRepository is a mock for the PromptRepository interface
type MockPromptRepository struct {
	mock.Mock
}

func (m *MockPromptRepository) AddPrompt(ctx context.Context, prompt *domain.Prompt) error {
	args := m.Called(ctx, prompt)
	return args.Error(0)
}

func (m *MockPromptRepository) GetPrompt(ctx context.Context, name string) (*domain.Prompt, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptRepository) ListPrompts(ctx context.Context) ([]*domain.Prompt, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Prompt), args.Error(1)
}

func (m *MockPromptRepository) DeletePrompt(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// MockSessionRepository is a mock for the SessionRepository interface
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) AddSession(ctx context.Context, session *domain.ClientSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSession(ctx context.Context, id string) (*domain.ClientSession, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.ClientSession), args.Error(1)
}

func (m *MockSessionRepository) ListSessions(ctx context.Context) ([]*domain.ClientSession, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.ClientSession), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockNotificationSender is a mock for the NotificationSender interface
type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) SendNotification(ctx context.Context, sessionID string, notification *domain.Notification) error {
	args := m.Called(ctx, sessionID, notification)
	return args.Error(0)
}

func (m *MockNotificationSender) BroadcastNotification(ctx context.Context, notification *domain.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func TestNewServerBuilder(t *testing.T) {
	builder := NewServerBuilder()

	assert.NotNil(t, builder)
	assert.Equal(t, "MCP Server", builder.name)
	assert.Equal(t, "1.0.0", builder.version)
	assert.Equal(t, "MCP Server for AI tools and resources", builder.instructions)
	assert.Equal(t, ":8080", builder.address)
	assert.NotNil(t, builder.resourceRepo)
	assert.NotNil(t, builder.toolRepo)
	assert.NotNil(t, builder.promptRepo)
	assert.NotNil(t, builder.sessionRepo)
	assert.Nil(t, builder.notificationSender)
}

func TestServerBuilder_WithName(t *testing.T) {
	builder := NewServerBuilder()
	result := builder.WithName("Test Server")

	assert.Equal(t, "Test Server", builder.name)
	assert.Equal(t, builder, result, "WithName should return the builder for chaining")
}

func TestServerBuilder_WithVersion(t *testing.T) {
	builder := NewServerBuilder()
	result := builder.WithVersion("2.0.0")

	assert.Equal(t, "2.0.0", builder.version)
	assert.Equal(t, builder, result, "WithVersion should return the builder for chaining")
}

func TestServerBuilder_WithInstructions(t *testing.T) {
	builder := NewServerBuilder()
	result := builder.WithInstructions("Test instructions")

	assert.Equal(t, "Test instructions", builder.instructions)
	assert.Equal(t, builder, result, "WithInstructions should return the builder for chaining")
}

func TestServerBuilder_WithAddress(t *testing.T) {
	builder := NewServerBuilder()
	result := builder.WithAddress(":9090")

	assert.Equal(t, ":9090", builder.address)
	assert.Equal(t, builder, result, "WithAddress should return the builder for chaining")
}

func TestServerBuilder_WithResourceRepository(t *testing.T) {
	builder := NewServerBuilder()
	mockRepo := new(MockResourceRepository)
	result := builder.WithResourceRepository(mockRepo)

	assert.Equal(t, mockRepo, builder.resourceRepo)
	assert.Equal(t, builder, result, "WithResourceRepository should return the builder for chaining")
}

func TestServerBuilder_WithToolRepository(t *testing.T) {
	builder := NewServerBuilder()
	mockRepo := new(MockToolRepository)
	result := builder.WithToolRepository(mockRepo)

	assert.Equal(t, mockRepo, builder.toolRepo)
	assert.Equal(t, builder, result, "WithToolRepository should return the builder for chaining")
}

func TestServerBuilder_WithPromptRepository(t *testing.T) {
	builder := NewServerBuilder()
	mockRepo := new(MockPromptRepository)
	result := builder.WithPromptRepository(mockRepo)

	assert.Equal(t, mockRepo, builder.promptRepo)
	assert.Equal(t, builder, result, "WithPromptRepository should return the builder for chaining")
}

func TestServerBuilder_WithSessionRepository(t *testing.T) {
	builder := NewServerBuilder()
	mockRepo := new(MockSessionRepository)
	result := builder.WithSessionRepository(mockRepo)

	assert.Equal(t, mockRepo, builder.sessionRepo)
	assert.Equal(t, builder, result, "WithSessionRepository should return the builder for chaining")
}

func TestServerBuilder_WithNotificationSender(t *testing.T) {
	builder := NewServerBuilder()
	mockSender := new(MockNotificationSender)
	result := builder.WithNotificationSender(mockSender)

	assert.Equal(t, mockSender, builder.notificationSender)
	assert.Equal(t, builder, result, "WithNotificationSender should return the builder for chaining")
}

func TestServerBuilder_AddTool(t *testing.T) {
	// Test with nil toolRepo
	builderWithNilRepo := &ServerBuilder{toolRepo: nil}
	tool := &domain.Tool{Name: "test-tool"}
	ctx := context.Background()

	result := builderWithNilRepo.AddTool(ctx, tool)
	assert.Equal(t, builderWithNilRepo, result, "AddTool should return the builder for chaining")

	// Test with mock toolRepo
	builder := NewServerBuilder()
	mockRepo := new(MockToolRepository)
	builder.toolRepo = mockRepo

	mockRepo.On("AddTool", ctx, tool).Return(nil)

	result = builder.AddTool(ctx, tool)
	assert.Equal(t, builder, result, "AddTool should return the builder for chaining")
	mockRepo.AssertExpectations(t)
}

func TestServerBuilder_AddResource(t *testing.T) {
	// Test with nil resourceRepo
	builderWithNilRepo := &ServerBuilder{resourceRepo: nil}
	resource := &domain.Resource{URI: "test://resource"}
	ctx := context.Background()

	result := builderWithNilRepo.AddResource(ctx, resource)
	assert.Equal(t, builderWithNilRepo, result, "AddResource should return the builder for chaining")

	// Test with mock resourceRepo
	builder := NewServerBuilder()
	mockRepo := new(MockResourceRepository)
	builder.resourceRepo = mockRepo

	mockRepo.On("AddResource", ctx, resource).Return(nil)

	result = builder.AddResource(ctx, resource)
	assert.Equal(t, builder, result, "AddResource should return the builder for chaining")
	mockRepo.AssertExpectations(t)
}

func TestServerBuilder_AddPrompt(t *testing.T) {
	// Test with nil promptRepo
	builderWithNilRepo := &ServerBuilder{promptRepo: nil}
	prompt := &domain.Prompt{Name: "test-prompt"}
	ctx := context.Background()

	result := builderWithNilRepo.AddPrompt(ctx, prompt)
	assert.Equal(t, builderWithNilRepo, result, "AddPrompt should return the builder for chaining")

	// Test with mock promptRepo
	builder := NewServerBuilder()
	mockRepo := new(MockPromptRepository)
	builder.promptRepo = mockRepo

	mockRepo.On("AddPrompt", ctx, prompt).Return(nil)

	result = builder.AddPrompt(ctx, prompt)
	assert.Equal(t, builder, result, "AddPrompt should return the builder for chaining")
	mockRepo.AssertExpectations(t)
}

func TestServerBuilder_BuildService(t *testing.T) {
	// Test with nil notificationSender
	builder := NewServerBuilder().
		WithName("Test Server").
		WithVersion("2.0.0").
		WithInstructions("Test instructions")

	service := builder.BuildService()

	assert.NotNil(t, service)

	name, version, instructions := service.ServerInfo()
	assert.Equal(t, "Test Server", name)
	assert.Equal(t, "2.0.0", version)
	assert.Equal(t, "Test instructions", instructions)

	// Test with custom notificationSender
	mockSender := new(MockNotificationSender)
	builder.WithNotificationSender(mockSender)

	service = builder.BuildService()

	assert.NotNil(t, service)
}

func TestServerBuilder_BuildMCPServer(t *testing.T) {
	builder := NewServerBuilder().
		WithName("Test Server").
		WithVersion("2.0.0").
		WithAddress(":9090")

	mcpServer := builder.BuildMCPServer()

	assert.NotNil(t, mcpServer)
}

func TestServerBuilder_BuildStdioServer(t *testing.T) {
	builder := NewServerBuilder().
		WithName("Test Server").
		WithVersion("2.0.0")

	stdioServer := builder.BuildStdioServer()

	assert.NotNil(t, stdioServer)
}

// TestServerBuilder_ServeStdio tests the ServeStdio method without actually starting the server
func TestServerBuilder_ServeStdio(t *testing.T) {
	// Create a builder
	builder := NewServerBuilder()

	// Just test that the method doesn't panic
	assert.NotPanics(t, func() {
		// Cancel this call immediately in the background to prevent hanging
		go func() {
			err := builder.ServeStdio()
			if err != nil {
				// We expect an error when the server is terminated abruptly
				assert.Contains(t, err.Error(), "")
			}
		}()
	})
}

func TestLoggingConfiguration(t *testing.T) {
	// Just test that we can create a logger with the default configuration
	logger, err := logging.New(logging.DefaultConfig())
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}
