package pocketbase

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreePeak/cortex/pkg/tools"
)

func TestNewCortexPlugin(t *testing.T) {
	// Create a plugin with default options
	plugin := NewCortexPlugin()

	// Check default values
	assert.Equal(t, "Cortex MCP Server", plugin.name)
	assert.Equal(t, "1.0.0", plugin.version)
	assert.Equal(t, "/api/mcp", plugin.basePath)
	assert.NotNil(t, plugin.logger)
	assert.NotNil(t, plugin.mcpServer)
}

func TestWithOptions(t *testing.T) {
	// Create a custom logger
	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)

	// Create a plugin with custom options
	plugin := NewCortexPlugin(
		WithName("Test Server"),
		WithVersion("2.0.0"),
		WithBasePath("/test"),
		WithLogger(logger),
	)

	// Check that options were applied
	assert.Equal(t, "Test Server", plugin.name)
	assert.Equal(t, "2.0.0", plugin.version)
	assert.Equal(t, "/test", plugin.basePath)
	assert.Equal(t, logger, plugin.logger)
}

func TestAddTool(t *testing.T) {
	// Create a plugin
	plugin := NewCortexPlugin()

	// Create a tool
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Add the tool
	err := plugin.AddTool(echoTool, func(ctx context.Context, request ToolCallRequest) (interface{}, error) {
		message := request.Parameters["message"].(string)
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		}, nil
	})

	// Check that there was no error
	assert.NoError(t, err)
}

func TestGetServerInfo(t *testing.T) {
	// Create a plugin with custom name and version
	plugin := NewCortexPlugin(
		WithName("Test Server"),
		WithVersion("2.0.0"),
	)

	// Get server info
	info := plugin.GetServerInfo()

	// Check info values
	assert.Equal(t, "Test Server", info.Name)
	assert.Equal(t, "2.0.0", info.Version)
	assert.NotEmpty(t, info.Address)
}
