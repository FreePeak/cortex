package builder

import (
	"context"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
)

// ToolHandlerFunc is an adapter function for handling tool calls in the MCP server.
// This wraps a custom handler function in the format expected by the internal APIs.
type ToolHandlerFunc func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error)

// WithToolHandler returns a stdio option to handle a specific tool.
// This allows you to register custom handlers for tools.
func WithToolHandler(toolName string, handler ToolHandlerFunc) stdio.StdioOption {
	return stdio.WithToolHandler(toolName, handler)
}
