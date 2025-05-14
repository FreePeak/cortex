// Package server provides the MCP server implementation.
package server

import (
	"context"
	"net/http"

	"github.com/FreePeak/cortex/pkg/types"
)

// ServerInfo contains basic information about the MCP server
type ServerInfo struct {
	// Name is the name of the server
	Name string

	// Version is the version of the server
	Version string

	// Address is the address the server is listening on
	Address string
}

// Embeddable defines an interface for MCP servers that can be embedded
// in other applications like PocketBase, standard HTTP servers, etc.
type Embeddable interface {
	// ToHTTPHandler returns an http.Handler that can be used to integrate
	// the MCP server with any standard Go HTTP server
	ToHTTPHandler() http.Handler

	// AddTool adds a tool to the MCP server
	AddTool(ctx context.Context, tool *types.Tool, handler ToolHandler) error

	// GetServerInfo returns basic information about the server
	GetServerInfo() ServerInfo

	// RegisterSession registers a new client session with the server
	// This allows sending messages to specific clients
	RegisterSession(sessionID string, userAgent string, callback func([]byte) error) error

	// UnregisterSession removes a client session from the server
	UnregisterSession(sessionID string) error

	// ExecuteTool executes a tool with the given request
	ExecuteTool(ctx context.Context, request ToolCallRequest) (interface{}, error)

	// SendToSession sends a message to a specific session
	SendToSession(sessionID string, message []byte) error

	// GetTools returns all registered tools
	GetTools() map[string]*types.Tool
}

// Ensure MCPServer implements Embeddable
var _ Embeddable = (*MCPServer)(nil)

// SetEmbedMode configures the MCPServer to work in embedded mode,
// which prevents it from binding to a port when ToHTTPHandler is called.
func (s *MCPServer) SetEmbedMode(embed bool) {
	s.embedMode = embed
}

// ToHTTPHandler returns an http.Handler for the MCP server
func (s *MCPServer) ToHTTPHandler() http.Handler {
	// When in embed mode, we provide a simpler handler that doesn't try to bind to a port
	if s.embedMode {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple server info endpoint
			if r.URL.Path == "/sse" || r.URL.Path == "/events" {
				// For SSE endpoints, send a simple event stream response
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")

				// Send the server info - using JSON-RPC 2.0 format
				serverInfo := `{"jsonrpc":"2.0","result":{"name":"` + s.name + `","version":"` + s.version + `","status":"running"},"id":"server.info"}`
				w.Write([]byte("event: server\ndata: " + serverInfo + "\n\n"))

				// Keep the connection open
				<-r.Context().Done()
				return
			}

			if r.URL.Path == "/status" || r.URL.Path == "/" {
				// Return JSON-RPC 2.0 formatted server info for status requests
				w.Header().Set("Content-Type", "application/json")
				jsonResponse := `{"jsonrpc":"2.0","result":{"name":"` + s.name + `","version":"` + s.version + `","status":"running"},"id":"server.info"}`
				w.Write([]byte(jsonResponse))
				return
			}

			// For all other requests, return a 404 for now - we need a more sophisticated solution
			http.NotFound(w, r)
		})
	}

	// If not in embed mode, use the default implementation (which may try to start the server)
	// Create a handler that delegates to the MCP server's internal handlers without starting the server
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pass the request to the MCP server's handlers
		// This is a fallback but will likely not work correctly since the server isn't started
		http.Error(w, "The MCP server is not configured for embedding", http.StatusInternalServerError)
	})
}

// GetServerInfo returns basic information about the server
func (s *MCPServer) GetServerInfo() ServerInfo {
	// GetAddress() may return the default value (:8080), but we want to provide
	// the address that reflects what the server will actually use when embedded
	return ServerInfo{
		Name:    s.name,
		Version: s.version,
		Address: s.GetAddress(),
	}
}
