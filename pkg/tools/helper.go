// Package tools provides utility functions for creating MCP tools.
package tools

import (
	"github.com/FreePeak/golang-mcp-server-sdk/pkg/types"
)

// ToolOption is a function that configures a tool.
type ToolOption func(*types.Tool)

// NewTool creates a new MCP tool with the given name and options.
func NewTool(name string, options ...ToolOption) *types.Tool {
	tool := &types.Tool{
		Name:       name,
		Parameters: []types.ToolParameter{},
	}

	// Apply all options
	for _, option := range options {
		option(tool)
	}

	return tool
}

// WithDescription sets the description of a tool.
func WithDescription(description string) ToolOption {
	return func(t *types.Tool) {
		t.Description = description
	}
}

// Parameter types

// ParameterOption is a function that configures a parameter.
type ParameterOption func(*types.ToolParameter)

// Description sets the description of a parameter.
func Description(description string) ParameterOption {
	return func(p *types.ToolParameter) {
		p.Description = description
	}
}

// Required marks a parameter as required.
func Required() ParameterOption {
	return func(p *types.ToolParameter) {
		p.Required = true
	}
}

// Type functions for creating parameters

// WithString adds a string parameter to a tool.
func WithString(name string, options ...ParameterOption) ToolOption {
	return func(t *types.Tool) {
		param := types.ToolParameter{
			Name: name,
			Type: "string",
		}

		// Apply options
		for _, option := range options {
			option(&param)
		}

		t.Parameters = append(t.Parameters, param)
	}
}

// WithNumber adds a number parameter to a tool.
func WithNumber(name string, options ...ParameterOption) ToolOption {
	return func(t *types.Tool) {
		param := types.ToolParameter{
			Name: name,
			Type: "number",
		}

		// Apply options
		for _, option := range options {
			option(&param)
		}

		t.Parameters = append(t.Parameters, param)
	}
}

// WithBoolean adds a boolean parameter to a tool.
func WithBoolean(name string, options ...ParameterOption) ToolOption {
	return func(t *types.Tool) {
		param := types.ToolParameter{
			Name: name,
			Type: "boolean",
		}

		// Apply options
		for _, option := range options {
			option(&param)
		}

		t.Parameters = append(t.Parameters, param)
	}
}

// WithArray adds an array parameter to a tool.
func WithArray(name string, options ...ParameterOption) ToolOption {
	return func(t *types.Tool) {
		param := types.ToolParameter{
			Name: name,
			Type: "array",
		}

		// Apply options
		for _, option := range options {
			option(&param)
		}

		t.Parameters = append(t.Parameters, param)
	}
}

// WithObject adds an object parameter to a tool.
func WithObject(name string, options ...ParameterOption) ToolOption {
	return func(t *types.Tool) {
		param := types.ToolParameter{
			Name: name,
			Type: "object",
		}

		// Apply options
		for _, option := range options {
			option(&param)
		}

		t.Parameters = append(t.Parameters, param)
	}
}
