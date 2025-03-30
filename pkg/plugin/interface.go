// Package plugin defines interfaces and utilities for the Cortex plugin system.
package plugin

import (
	"context"

	"github.com/FreePeak/cortex/pkg/types"
)

// Provider represents a tool provider that can register tools with the Cortex platform.
type Provider interface {
	// GetProviderInfo returns information about the tool provider.
	GetProviderInfo(ctx context.Context) (*ProviderInfo, error)

	// GetTools returns a list of tools provided by this provider.
	GetTools(ctx context.Context) ([]*types.Tool, error)

	// ExecuteTool executes a specific tool with the given parameters.
	ExecuteTool(ctx context.Context, request *ExecuteRequest) (*ExecuteResponse, error)
}

// ProviderInfo contains metadata about a tool provider.
type ProviderInfo struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	URL         string
}

// ExecuteRequest contains information for executing a tool.
type ExecuteRequest struct {
	ToolName   string
	Parameters map[string]interface{}
	Session    *types.ClientSession
}

// ExecuteResponse contains the result of executing a tool.
type ExecuteResponse struct {
	Content interface{}
	Error   error
}

// Registry manages the registration and discovery of tool providers.
type Registry interface {
	// RegisterProvider registers a new tool provider with the registry.
	RegisterProvider(ctx context.Context, provider Provider) error

	// UnregisterProvider removes a tool provider from the registry.
	UnregisterProvider(ctx context.Context, providerID string) error

	// GetProvider retrieves a specific provider by ID.
	GetProvider(ctx context.Context, providerID string) (Provider, error)

	// ListProviders returns all registered providers.
	ListProviders(ctx context.Context) ([]Provider, error)

	// GetTool retrieves a specific tool by name.
	GetTool(ctx context.Context, toolName string) (*types.Tool, Provider, error)

	// ListTools returns all tools from all registered providers.
	ListTools(ctx context.Context) ([]*types.Tool, error)
}
