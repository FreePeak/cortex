package stdio

import (
	"context"
	"log"
	"os"

	"github.com/FreePeak/cortex/internal/domain"
)

// WithToolHandler registers a custom handler function for a specific tool.
// This allows you to override the default tool handling behavior.
func WithToolHandler(toolName string, handler func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error)) StdioOption {
	return func(s *StdioServer) {
		debugMode := os.Getenv("CORTEX_DEBUG") == "1"

		if s.processor == nil {
			s.processor = NewMessageProcessor(s.server, s.logger)
		}

		// Initialize toolHandlers map if needed
		if s.processor.toolHandlers == nil {
			s.processor.toolHandlers = make(map[string]func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error))
		}

		// Store the handler in the toolHandlers map
		s.processor.toolHandlers[toolName] = handler

		if debugMode {
			log.Printf("Registered handler for tool: %s", toolName)
		}

		// Register a single handler for tools/call if not already done
		if _, exists := s.processor.handlers["tools/call"]; !exists {
			s.processor.RegisterHandler("tools/call", MethodHandlerFunc(s.processor.handleToolsCall))
		}
	}
}

// WithAllToolHandlers sets all tool handlers at once.
// This is useful when you want to set multiple handlers in a single operation.
func WithAllToolHandlers(handlers map[string]func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error)) StdioOption {
	return func(s *StdioServer) {
		debugMode := os.Getenv("CORTEX_DEBUG") == "1"

		if s.processor == nil {
			s.processor = NewMessageProcessor(s.server, s.logger)
		}

		// Initialize toolHandlers map if needed
		if s.processor.toolHandlers == nil {
			s.processor.toolHandlers = make(map[string]func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error))
		}

		// Copy all handlers to the toolHandlers map
		for name, handler := range handlers {
			s.processor.toolHandlers[name] = handler
		}

		if debugMode {
			log.Printf("Registered %d tool handlers at once", len(handlers))
			for name := range handlers {
				log.Printf("Tool handler registered: %s", name)
			}
		}

		// Register a single handler for tools/call if not already done
		if _, exists := s.processor.handlers["tools/call"]; !exists {
			s.processor.RegisterHandler("tools/call", MethodHandlerFunc(s.processor.handleToolsCall))
		}
	}
}
