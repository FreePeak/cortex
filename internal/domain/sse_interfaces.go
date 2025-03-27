package domain

import (
	"context"
	"encoding/json"
	"net/http"
)

// SSEHandler defines the interface for a server-sent events handler.
type SSEHandler interface {
	// ServeHTTP handles HTTP requests for SSE events.
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// Start starts the SSE server.
	Start(addr string) error

	// Shutdown gracefully stops the SSE server.
	Shutdown(ctx context.Context) error

	// BroadcastEvent sends an event to all connected clients.
	BroadcastEvent(event interface{}) error

	// SendEventToSession sends an event to a specific client session.
	SendEventToSession(sessionID string, event interface{}) error
}

// MessageHandler defines the interface for handling message processing in the SSE server.
type MessageHandler interface {
	// HandleMessage processes a raw JSON message and returns a response.
	HandleMessage(ctx context.Context, rawMessage json.RawMessage) interface{}
}

// SSESession represents an active SSE connection.
type SSESession interface {
	// ID returns the session identifier.
	ID() string

	// Close closes the session.
	Close()

	// NotificationChannel returns the channel used to send notifications to this session.
	NotificationChannel() chan<- string

	// Start begins processing events for this session.
	Start()

	// Context returns the session's context.
	Context() context.Context
}

// ConnectionManager defines the interface for managing SSE connections.
type ConnectionManager interface {
	// Add adds a session to the manager.
	AddSession(session SSESession)

	// Remove removes a session from the manager.
	RemoveSession(sessionID string)

	// Get retrieves a session by ID.
	GetSession(sessionID string) (SSESession, bool)

	// Broadcast sends an event to all sessions.
	Broadcast(event interface{}) error

	// CloseAll closes all active sessions.
	CloseAll()

	// Count returns the number of active sessions.
	Count() int
}
