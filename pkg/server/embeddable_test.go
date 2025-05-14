package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/FreePeak/cortex/pkg/types"
)

// MockEmbeddable is a mock implementation of the Embeddable interface for testing
type MockEmbeddable struct {
	mock.Mock
}

func (m *MockEmbeddable) ToHTTPHandler() http.Handler {
	args := m.Called()
	return args.Get(0).(http.Handler)
}

func (m *MockEmbeddable) AddTool(ctx context.Context, tool *types.Tool, handler ToolHandler) error {
	args := m.Called(ctx, tool, handler)
	return args.Error(0)
}

func (m *MockEmbeddable) GetServerInfo() ServerInfo {
	args := m.Called()
	return args.Get(0).(ServerInfo)
}

func TestEmbeddableInterface(t *testing.T) {
	// This test verifies that MCPServer implements the Embeddable interface
	var _ Embeddable = (*MCPServer)(nil)
}

func TestToHTTPHandler(t *testing.T) {
	mockEmb := new(MockEmbeddable)
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mockEmb.On("ToHTTPHandler").Return(mockHandler)

	handler := mockEmb.ToHTTPHandler()

	assert.NotNil(t, handler)
	mockEmb.AssertExpectations(t)
}

func TestAddTool(t *testing.T) {
	mockEmb := new(MockEmbeddable)
	ctx := context.Background()
	mockTool := &types.Tool{
		Name:        "test-tool",
		Description: "Test tool",
	}

	mockHandler := func(ctx context.Context, request ToolCallRequest) (interface{}, error) {
		return nil, nil
	}

	mockEmb.On("AddTool", ctx, mockTool, mock.AnythingOfType("ToolHandler")).Return(nil)

	err := mockEmb.AddTool(ctx, mockTool, mockHandler)

	assert.NoError(t, err)
	mockEmb.AssertExpectations(t)
}

func TestGetServerInfo(t *testing.T) {
	mockEmb := new(MockEmbeddable)
	expectedInfo := ServerInfo{
		Name:    "Test Server",
		Version: "1.0.0",
		Address: "localhost:8080",
	}

	mockEmb.On("GetServerInfo").Return(expectedInfo)

	info := mockEmb.GetServerInfo()

	assert.Equal(t, expectedInfo, info)
	mockEmb.AssertExpectations(t)
}
