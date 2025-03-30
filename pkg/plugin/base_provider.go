package plugin

import (
	"context"
	"fmt"
	"log"

	"github.com/FreePeak/cortex/pkg/types"
)

// ToolExecutor is a function that executes a tool and returns a result.
type ToolExecutor func(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error)

// BaseProvider implements the Provider interface and provides a foundation for building tool providers.
type BaseProvider struct {
	info      ProviderInfo
	tools     []*types.Tool
	executors map[string]ToolExecutor
	logger    *log.Logger
}

// NewBaseProvider creates a new BaseProvider with the given info.
func NewBaseProvider(info ProviderInfo, logger *log.Logger) *BaseProvider {
	if logger == nil {
		logger = log.Default()
	}

	return &BaseProvider{
		info:      info,
		tools:     make([]*types.Tool, 0),
		executors: make(map[string]ToolExecutor),
		logger:    logger,
	}
}

// GetProviderInfo returns information about the tool provider.
func (p *BaseProvider) GetProviderInfo(ctx context.Context) (*ProviderInfo, error) {
	return &p.info, nil
}

// GetTools returns a list of tools provided by this provider.
func (p *BaseProvider) GetTools(ctx context.Context) ([]*types.Tool, error) {
	return p.tools, nil
}

// ExecuteTool executes a specific tool with the given parameters.
func (p *BaseProvider) ExecuteTool(ctx context.Context, request *ExecuteRequest) (*ExecuteResponse, error) {
	// Validate the request
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if request.ToolName == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	// Get the executor for the tool
	executor, exists := p.executors[request.ToolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", request.ToolName)
	}

	// Execute the tool
	p.logger.Printf("Executing tool: %s", request.ToolName)
	result, err := executor(ctx, request.Parameters, request.Session)
	if err != nil {
		p.logger.Printf("Error executing tool %s: %v", request.ToolName, err)
		return &ExecuteResponse{Error: err}, nil
	}

	// Return the result
	return &ExecuteResponse{Content: result}, nil
}

// RegisterTool registers a new tool with the provider.
func (p *BaseProvider) RegisterTool(tool *types.Tool, executor ToolExecutor) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	if executor == nil {
		return fmt.Errorf("executor cannot be nil")
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	// Check if the tool already exists
	for _, existingTool := range p.tools {
		if existingTool.Name == tool.Name {
			return fmt.Errorf("tool %s already registered", tool.Name)
		}
	}

	// Add the tool and its executor
	p.tools = append(p.tools, tool)
	p.executors[tool.Name] = executor
	p.logger.Printf("Registered tool %s with provider %s", tool.Name, p.info.ID)

	return nil
}

// UnregisterTool removes a tool from the provider.
func (p *BaseProvider) UnregisterTool(toolName string) error {
	// Find the tool
	index := -1
	for i, tool := range p.tools {
		if tool.Name == toolName {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("tool %s not found", toolName)
	}

	// Remove the tool
	p.tools = append(p.tools[:index], p.tools[index+1:]...)
	delete(p.executors, toolName)
	p.logger.Printf("Unregistered tool %s from provider %s", toolName, p.info.ID)

	return nil
}
