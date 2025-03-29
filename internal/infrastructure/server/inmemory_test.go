package server

import (
	"context"
	"testing"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryResourceRepository(t *testing.T) {
	repo := NewInMemoryResourceRepository()
	assert.NotNil(t, repo)
}

func TestInMemoryResourceRepository_AddAndGetResource(t *testing.T) {
	repo := NewInMemoryResourceRepository()
	ctx := context.Background()

	// Create a test resource
	resource := &domain.Resource{
		URI:         "test://resource1",
		Name:        "Test Resource",
		Description: "A test resource",
	}

	// Add the resource
	err := repo.AddResource(ctx, resource)
	require.NoError(t, err)

	// Get the resource
	retrieved, err := repo.GetResource(ctx, "test://resource1")
	assert.NoError(t, err)
	assert.Equal(t, resource, retrieved)

	// Try to get a non-existent resource
	_, err = repo.GetResource(ctx, "test://nonexistent")
	assert.Error(t, err)
	// Check if it's the right error type
	_, ok := err.(*domain.ResourceNotFoundError)
	assert.True(t, ok)
}

func TestInMemoryResourceRepository_ListResources(t *testing.T) {
	repo := NewInMemoryResourceRepository()
	ctx := context.Background()

	// Create test resources
	resource1 := &domain.Resource{URI: "test://resource1", Name: "Resource 1"}
	resource2 := &domain.Resource{URI: "test://resource2", Name: "Resource 2"}

	// Add resources
	err := repo.AddResource(ctx, resource1)
	require.NoError(t, err)
	err = repo.AddResource(ctx, resource2)
	require.NoError(t, err)

	// List resources
	resources, err := repo.ListResources(ctx)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)

	// Verify the resources are in the list
	uris := []string{resources[0].URI, resources[1].URI}
	assert.Contains(t, uris, "test://resource1")
	assert.Contains(t, uris, "test://resource2")
}

func TestInMemoryResourceRepository_DeleteResource(t *testing.T) {
	repo := NewInMemoryResourceRepository()
	ctx := context.Background()

	// Create and add a resource
	resource := &domain.Resource{URI: "test://resource1", Name: "Test Resource"}
	err := repo.AddResource(ctx, resource)
	require.NoError(t, err)

	// Delete the resource
	err = repo.DeleteResource(ctx, "test://resource1")
	assert.NoError(t, err)

	// Verify the resource was deleted
	_, err = repo.GetResource(ctx, "test://resource1")
	assert.Error(t, err)

	// Try to delete a non-existent resource
	err = repo.DeleteResource(ctx, "test://nonexistent")
	assert.Error(t, err)
	_, ok := err.(*domain.ResourceNotFoundError)
	assert.True(t, ok)
}

func TestNewInMemoryToolRepository(t *testing.T) {
	repo := NewInMemoryToolRepository()
	assert.NotNil(t, repo)
}

func TestInMemoryToolRepository_AddAndGetTool(t *testing.T) {
	repo := NewInMemoryToolRepository()
	ctx := context.Background()

	// Create a test tool
	tool := &domain.Tool{
		Name:        "test-tool",
		Description: "A test tool",
	}

	// Add the tool
	err := repo.AddTool(ctx, tool)
	require.NoError(t, err)

	// Get the tool
	retrieved, err := repo.GetTool(ctx, "test-tool")
	assert.NoError(t, err)
	assert.Equal(t, tool, retrieved)

	// Try to get a non-existent tool
	_, err = repo.GetTool(ctx, "nonexistent-tool")
	assert.Error(t, err)
	_, ok := err.(*domain.ToolNotFoundError)
	assert.True(t, ok)
}

func TestInMemoryToolRepository_ListTools(t *testing.T) {
	repo := NewInMemoryToolRepository()
	ctx := context.Background()

	// Create test tools
	tool1 := &domain.Tool{Name: "tool1", Description: "Tool 1"}
	tool2 := &domain.Tool{Name: "tool2", Description: "Tool 2"}

	// Add tools
	err := repo.AddTool(ctx, tool1)
	require.NoError(t, err)
	err = repo.AddTool(ctx, tool2)
	require.NoError(t, err)

	// List tools
	tools, err := repo.ListTools(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 2)

	// Verify the tools are in the list
	names := []string{tools[0].Name, tools[1].Name}
	assert.Contains(t, names, "tool1")
	assert.Contains(t, names, "tool2")
}

func TestInMemoryToolRepository_DeleteTool(t *testing.T) {
	repo := NewInMemoryToolRepository()
	ctx := context.Background()

	// Create and add a tool
	tool := &domain.Tool{Name: "test-tool", Description: "Test Tool"}
	err := repo.AddTool(ctx, tool)
	require.NoError(t, err)

	// Delete the tool
	err = repo.DeleteTool(ctx, "test-tool")
	assert.NoError(t, err)

	// Verify the tool was deleted
	_, err = repo.GetTool(ctx, "test-tool")
	assert.Error(t, err)

	// Try to delete a non-existent tool
	err = repo.DeleteTool(ctx, "nonexistent-tool")
	assert.Error(t, err)
	_, ok := err.(*domain.ToolNotFoundError)
	assert.True(t, ok)
}

func TestNewInMemorySessionRepository(t *testing.T) {
	repo := NewInMemorySessionRepository()
	assert.NotNil(t, repo)
}

func TestInMemorySessionRepository_AddAndGetSession(t *testing.T) {
	repo := NewInMemorySessionRepository()
	ctx := context.Background()

	// Create a test session
	session := &domain.ClientSession{
		ID:        "session-1",
		UserAgent: "test-agent",
	}

	// Add the session
	err := repo.AddSession(ctx, session)
	require.NoError(t, err)

	// Get the session
	retrieved, err := repo.GetSession(ctx, "session-1")
	assert.NoError(t, err)
	assert.Equal(t, session, retrieved)

	// Try to get a non-existent session
	_, err = repo.GetSession(ctx, "nonexistent-session")
	assert.Error(t, err)
	_, ok := err.(*domain.SessionNotFoundError)
	assert.True(t, ok)
}

func TestInMemorySessionRepository_ListSessions(t *testing.T) {
	repo := NewInMemorySessionRepository()
	ctx := context.Background()

	// Create test sessions
	session1 := &domain.ClientSession{ID: "session-1", UserAgent: "Agent 1"}
	session2 := &domain.ClientSession{ID: "session-2", UserAgent: "Agent 2"}

	// Add sessions
	err := repo.AddSession(ctx, session1)
	require.NoError(t, err)
	err = repo.AddSession(ctx, session2)
	require.NoError(t, err)

	// List sessions
	sessions, err := repo.ListSessions(ctx)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Verify the sessions are in the list
	ids := []string{sessions[0].ID, sessions[1].ID}
	assert.Contains(t, ids, "session-1")
	assert.Contains(t, ids, "session-2")
}

func TestInMemorySessionRepository_DeleteSession(t *testing.T) {
	repo := NewInMemorySessionRepository()
	ctx := context.Background()

	// Create and add a session
	session := &domain.ClientSession{ID: "session-1", UserAgent: "Test Agent"}
	err := repo.AddSession(ctx, session)
	require.NoError(t, err)

	// Delete the session
	err = repo.DeleteSession(ctx, "session-1")
	assert.NoError(t, err)

	// Verify the session was deleted
	_, err = repo.GetSession(ctx, "session-1")
	assert.Error(t, err)

	// Try to delete a non-existent session
	err = repo.DeleteSession(ctx, "nonexistent-session")
	assert.Error(t, err)
	_, ok := err.(*domain.SessionNotFoundError)
	assert.True(t, ok)
}
