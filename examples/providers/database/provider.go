// Package database provides a database service provider for the Cortex platform.
package database

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/FreePeak/cortex/pkg/plugin"
	"github.com/FreePeak/cortex/pkg/tools"
	"github.com/FreePeak/cortex/pkg/types"
)

// SimpleDatabase is an in-memory key-value store for the example.
type SimpleDatabase struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSimpleDatabase creates a new in-memory database.
func NewSimpleDatabase() *SimpleDatabase {
	return &SimpleDatabase{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the database.
func (db *SimpleDatabase) Set(key string, value interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[key] = value
}

// Get retrieves a value from the database.
func (db *SimpleDatabase) Get(key string) (interface{}, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	value, exists := db.data[key]
	return value, exists
}

// Delete removes a value from the database.
func (db *SimpleDatabase) Delete(key string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.data[key]; !exists {
		return false
	}

	delete(db.data, key)
	return true
}

// Keys returns all keys in the database.
func (db *SimpleDatabase) Keys() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keys := make([]string, 0, len(db.data))
	for k := range db.data {
		keys = append(keys, k)
	}

	return keys
}

// DBProvider implements the plugin.Provider interface for database operations.
type DBProvider struct {
	*plugin.BaseProvider
	db *SimpleDatabase
}

// NewDBProvider creates a new database provider.
func NewDBProvider(logger *log.Logger) (*DBProvider, error) {
	// Create provider info
	info := plugin.ProviderInfo{
		ID:          "cortex-db-provider",
		Name:        "Database Provider",
		Version:     "1.0.0",
		Description: "A provider for simple database operations",
		Author:      "Cortex Team",
		URL:         "https://github.com/FreePeak/cortex",
	}

	// Create base provider
	baseProvider := plugin.NewBaseProvider(info, logger)

	// Create database instance
	db := NewSimpleDatabase()

	// Create database provider
	provider := &DBProvider{
		BaseProvider: baseProvider,
		db:           db,
	}

	// Register database tools
	err := provider.registerTools()
	if err != nil {
		return nil, fmt.Errorf("failed to register database tools: %w", err)
	}

	return provider, nil
}

// registerTools registers all database tools with the provider.
func (p *DBProvider) registerTools() error {
	// Register get tool
	getTool := tools.NewTool("db.get",
		tools.WithDescription("Gets a value from the database"),
		tools.WithString("key",
			tools.Description("The key to retrieve"),
			tools.Required(),
		),
	)
	err := p.RegisterTool(getTool, p.handleGet)
	if err != nil {
		return fmt.Errorf("failed to register get tool: %w", err)
	}

	// Register set tool
	setTool := tools.NewTool("db.set",
		tools.WithDescription("Sets a value in the database"),
		tools.WithString("key",
			tools.Description("The key to set"),
			tools.Required(),
		),
		tools.WithObject("value",
			tools.Description("The value to store"),
			tools.Required(),
		),
	)
	err = p.RegisterTool(setTool, p.handleSet)
	if err != nil {
		return fmt.Errorf("failed to register set tool: %w", err)
	}

	// Register delete tool
	deleteTool := tools.NewTool("db.delete",
		tools.WithDescription("Deletes a value from the database"),
		tools.WithString("key",
			tools.Description("The key to delete"),
			tools.Required(),
		),
	)
	err = p.RegisterTool(deleteTool, p.handleDelete)
	if err != nil {
		return fmt.Errorf("failed to register delete tool: %w", err)
	}

	// Register keys tool
	keysTool := tools.NewTool("db.keys",
		tools.WithDescription("Lists all keys in the database"),
	)
	err = p.RegisterTool(keysTool, p.handleKeys)
	if err != nil {
		return fmt.Errorf("failed to register keys tool: %w", err)
	}

	return nil
}

// handleGet handles the db.get tool requests.
func (p *DBProvider) handleGet(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
	// Extract the key parameter
	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'key' parameter")
	}

	// Get the value from the database
	value, exists := p.db.Get(key)
	if !exists {
		return nil, fmt.Errorf("key '%s' not found", key)
	}

	// Return the value in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Value for key '%s': %v", key, value),
			},
		},
		"value": value,
	}, nil
}

// handleSet handles the db.set tool requests.
func (p *DBProvider) handleSet(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
	// Extract the key parameter
	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'key' parameter")
	}

	// Extract the value parameter
	value, ok := params["value"]
	if !ok {
		return nil, fmt.Errorf("missing 'value' parameter")
	}

	// Set the value in the database
	p.db.Set(key, value)

	// Return the success message in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Successfully set value for key '%s'", key),
			},
		},
		"success": true,
	}, nil
}

// handleDelete handles the db.delete tool requests.
func (p *DBProvider) handleDelete(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
	// Extract the key parameter
	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'key' parameter")
	}

	// Delete the value from the database
	success := p.db.Delete(key)

	// Return the result in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Delete operation for key '%s': %v", key, success),
			},
		},
		"success": success,
	}, nil
}

// handleKeys handles the db.keys tool requests.
func (p *DBProvider) handleKeys(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
	// Get all keys from the database
	keys := p.db.Keys()

	// Format the keys
	var keysText string
	if len(keys) == 0 {
		keysText = "No keys found in the database."
	} else {
		keysText = "Database keys:\n"
		for i, key := range keys {
			keysText += fmt.Sprintf("%d. %s\n", i+1, key)
		}
	}

	// Return the keys in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": keysText,
			},
		},
		"keys": keys,
	}, nil
}
