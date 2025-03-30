// Package tools provides the tools infrastructure for the Cortex MCP platform.
package tools

import (
	"context"

	"github.com/FreePeak/cortex/internal/domain"
	"github.com/FreePeak/cortex/internal/infrastructure/logging"
)

// ToolRegistry provides registration methods for tool handlers
type ToolRegistry interface {
	RegisterToolHandler(name string, handler func(ctx context.Context, params map[string]interface{}, session *domain.ClientSession) (interface{}, error))
}

// ToolProvider is an interface for a service that provides tools
type ToolProvider interface {
	// RegisterTool registers the tool with the provided registry
	RegisterTool(registerFunc func(string, func(context.Context, map[string]interface{}, *domain.ClientSession) (interface{}, error)))

	// GetToolDefinitions returns the tool definitions
	GetToolDefinitions() []*domain.Tool
}

// Manager manages all available tools
type Manager struct {
	providers []ToolProvider
	logger    *logging.Logger
}

// NewManager creates a new tools manager
func NewManager(logger *logging.Logger) *Manager {
	if logger == nil {
		logger = logging.Default()
	}

	return &Manager{
		providers: []ToolProvider{},
		logger:    logger,
	}
}

// RegisterProvider registers an external tool provider
func (m *Manager) RegisterProvider(provider ToolProvider) {
	m.providers = append(m.providers, provider)
}

// GetAllTools returns all tool definitions from all providers
func (m *Manager) GetAllTools() []*domain.Tool {
	var allTools []*domain.Tool

	for _, provider := range m.providers {
		allTools = append(allTools, provider.GetToolDefinitions()...)
	}

	return allTools
}
