package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/FreePeak/cortex/pkg/types"
)

// DefaultRegistry is the default implementation of the Registry interface.
type DefaultRegistry struct {
	providers map[string]Provider
	toolMap   map[string]string // Maps tool names to provider IDs
	mu        sync.RWMutex
	logger    *log.Logger
}

// NewRegistry creates a new registry for managing tool providers.
func NewRegistry(logger *log.Logger) *DefaultRegistry {
	if logger == nil {
		logger = log.Default()
	}

	return &DefaultRegistry{
		providers: make(map[string]Provider),
		toolMap:   make(map[string]string),
		logger:    logger,
	}
}

// RegisterProvider registers a new tool provider with the registry.
func (r *DefaultRegistry) RegisterProvider(ctx context.Context, provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	info, err := provider.GetProviderInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get provider info: %w", err)
	}

	if info.ID == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}

	// Register the provider
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if provider already exists
	if _, exists := r.providers[info.ID]; exists {
		return fmt.Errorf("provider with ID %s is already registered", info.ID)
	}

	// Add provider to registry
	r.providers[info.ID] = provider
	r.logger.Printf("Registered provider: %s (%s)", info.Name, info.ID)

	// Register all tools provided by this provider
	tools, err := provider.GetTools(ctx)
	if err != nil {
		// We registered the provider but failed to get tools
		// Let's keep the provider registered but log the error
		r.logger.Printf("Error getting tools from provider %s: %v", info.ID, err)
		return nil
	}

	// Register all tools with this provider
	for _, tool := range tools {
		if tool.Name == "" {
			r.logger.Printf("Skipping tool with empty name from provider %s", info.ID)
			continue
		}

		// Check for tool name collision
		if existingProvider, exists := r.toolMap[tool.Name]; exists {
			r.logger.Printf("Tool name collision: %s already registered by provider %s", tool.Name, existingProvider)
			continue
		}

		r.toolMap[tool.Name] = info.ID
		r.logger.Printf("Registered tool: %s from provider %s", tool.Name, info.ID)
	}

	return nil
}

// UnregisterProvider removes a tool provider from the registry.
func (r *DefaultRegistry) UnregisterProvider(ctx context.Context, providerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if provider exists
	if _, exists := r.providers[providerID]; !exists {
		return fmt.Errorf("provider with ID %s is not registered", providerID)
	}

	// Remove all tools associated with this provider
	var toolsToRemove []string
	for toolName, id := range r.toolMap {
		if id == providerID {
			toolsToRemove = append(toolsToRemove, toolName)
		}
	}

	for _, toolName := range toolsToRemove {
		delete(r.toolMap, toolName)
	}

	// Remove the provider
	delete(r.providers, providerID)
	r.logger.Printf("Unregistered provider: %s with %d tools", providerID, len(toolsToRemove))

	return nil
}

// GetProvider retrieves a specific provider by ID.
func (r *DefaultRegistry) GetProvider(ctx context.Context, providerID string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[providerID]
	if !exists {
		return nil, fmt.Errorf("provider with ID %s is not registered", providerID)
	}

	return provider, nil
}

// ListProviders returns all registered providers.
func (r *DefaultRegistry) ListProviders(ctx context.Context) ([]Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetTool retrieves a specific tool by name.
func (r *DefaultRegistry) GetTool(ctx context.Context, toolName string) (*types.Tool, Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find the provider for this tool
	providerID, exists := r.toolMap[toolName]
	if !exists {
		return nil, nil, fmt.Errorf("tool %s is not registered", toolName)
	}

	// Get the provider
	provider, exists := r.providers[providerID]
	if !exists {
		// This should not happen, but handle it anyway
		return nil, nil, fmt.Errorf("provider for tool %s is no longer registered", toolName)
	}

	// Get all tools from the provider
	tools, err := provider.GetTools(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tools from provider %s: %w", providerID, err)
	}

	// Find the specific tool
	for _, tool := range tools {
		if tool.Name == toolName {
			return tool, provider, nil
		}
	}

	// Tool was registered but not found in provider's tools
	return nil, nil, fmt.Errorf("tool %s is no longer provided by provider %s", toolName, providerID)
}

// ListTools returns all tools from all registered providers.
func (r *DefaultRegistry) ListTools(ctx context.Context) ([]*types.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allTools []*types.Tool

	for providerID, provider := range r.providers {
		tools, err := provider.GetTools(ctx)
		if err != nil {
			r.logger.Printf("Error getting tools from provider %s: %v", providerID, err)
			continue
		}

		allTools = append(allTools, tools...)
	}

	return allTools, nil
}
