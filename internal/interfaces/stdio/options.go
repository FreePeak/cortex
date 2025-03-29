package stdio

import (
	"context"
	"fmt"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

// WithToolHandler registers a custom handler function for a specific tool.
// This allows you to override the default tool handling behavior.
func WithToolHandler(toolName string, handler func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error)) StdioOption {
	return func(s *StdioServer) {
		if s.processor == nil {
			s.processor = NewMessageProcessor(s.server, s.logger)
		}

		// Create an adapter that converts our handler function to a MethodHandlerFunc
		adapter := MethodHandlerFunc(func(ctx context.Context, params interface{}, id interface{}) (interface{}, *domain.JSONRPCError) {
			// Extract tool parameters
			paramsMap, ok := params.(map[string]interface{})
			if !ok {
				return nil, &domain.JSONRPCError{
					Code:    InvalidParamsCode,
					Message: "Invalid params",
				}
			}

			// Check if this is a call to our specific tool
			nameParam, ok := paramsMap["name"].(string)
			if !ok || nameParam != toolName {
				// Let the default handler handle other tools
				return nil, &domain.JSONRPCError{
					Code:    MethodNotFoundCode,
					Message: fmt.Sprintf("Tool handler mismatch: expected %s, got %s", toolName, nameParam),
				}
			}

			// Get tool parameters - check both parameters and arguments fields
			toolParams, ok := paramsMap["parameters"].(map[string]interface{})
			if !ok {
				// Try arguments field if parameters is not available
				toolParams, ok = paramsMap["arguments"].(map[string]interface{})
				if !ok {
					toolParams = map[string]interface{}{}
				}
			}

			// Create a dummy session for now
			session := &domain.ClientSession{
				ID:        "stdio-session",
				UserAgent: "stdio-client",
				Connected: true,
			}

			// Call the handler
			result, err := handler(ctx, toolParams, session)
			if err != nil {
				return nil, &domain.JSONRPCError{
					Code:    InternalErrorCode,
					Message: err.Error(),
				}
			}

			return result, nil
		})

		// Override the tools/call handler with our custom one
		s.processor.RegisterHandler("tools/call", adapter)
	}
}
