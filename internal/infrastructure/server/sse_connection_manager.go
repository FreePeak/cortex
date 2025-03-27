package server

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

// sseConnectionManager implements the domain.ConnectionManager interface
// for managing SSE connections.
type sseConnectionManager struct {
	mu       sync.RWMutex
	sessions map[string]domain.SSESession
}

// NewSSEConnectionManager creates a new connection manager for SSE sessions.
func NewSSEConnectionManager() domain.ConnectionManager {
	return &sseConnectionManager{
		sessions: make(map[string]domain.SSESession),
	}
}

// AddSession adds a session to the connection manager.
func (m *sseConnectionManager) AddSession(session domain.SSESession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID()] = session
}

// RemoveSession removes a session from the connection manager.
func (m *sseConnectionManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// GetSession retrieves a session by its ID.
func (m *sseConnectionManager) GetSession(sessionID string) (domain.SSESession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	return session, ok
}

// Broadcast sends an event to all connected sessions.
func (m *sseConnectionManager) Broadcast(event interface{}) error {
	// Format the event as a SSE message
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	eventStr := fmt.Sprintf("event: message\ndata: %s\n\n", eventData)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		select {
		case session.NotificationChannel() <- eventStr:
			// Event sent successfully
		default:
			// Queue is full or closed, we continue anyway to avoid blocking
		}
	}

	return nil
}

// CloseAll closes all active sessions.
func (m *sseConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.Close()
	}

	// Clear the map
	m.sessions = make(map[string]domain.SSESession)
}

// Count returns the number of active sessions.
func (m *sseConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
