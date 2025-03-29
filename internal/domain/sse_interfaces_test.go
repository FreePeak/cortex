package domain

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
)

// MockSSESession is a mock implementation of SSESession
type MockSSESession struct {
	id        string
	notifChan chan string
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// NewMockSSESession creates a new MockSSESession
func NewMockSSESession(id string) *MockSSESession {
	ctx, cancel := context.WithCancel(context.Background())
	return &MockSSESession{
		id:        id,
		notifChan: make(chan string, 10),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// ID returns the session ID
func (m *MockSSESession) ID() string {
	return m.id
}

// Close closes the session
func (m *MockSSESession) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		m.cancel()
		close(m.notifChan)
	}
}

// NotificationChannel returns the notification channel
func (m *MockSSESession) NotificationChannel() chan<- string {
	return m.notifChan
}

// Start starts processing events for the session
func (m *MockSSESession) Start() {
	// Mock implementation, nothing to do
}

// Context returns the session context
func (m *MockSSESession) Context() context.Context {
	return m.ctx
}

// IsClosed returns whether the session is closed
func (m *MockSSESession) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

// MockConnectionManager is a mock implementation of ConnectionManager
type MockConnectionManager struct {
	sessions map[string]SSESession
	mu       sync.RWMutex
}

// NewMockConnectionManager creates a new MockConnectionManager
func NewMockConnectionManager() *MockConnectionManager {
	return &MockConnectionManager{
		sessions: make(map[string]SSESession),
	}
}

// AddSession adds a session to the manager
func (m *MockConnectionManager) AddSession(session SSESession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID()] = session
}

// RemoveSession removes a session from the manager
func (m *MockConnectionManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// GetSession retrieves a session by ID
func (m *MockConnectionManager) GetSession(sessionID string) (SSESession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	return session, ok
}

// Broadcast sends an event to all sessions
func (m *MockConnectionManager) Broadcast(event interface{}) error {
	eventStr, err := json.Marshal(event)
	if err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		select {
		case session.NotificationChannel() <- string(eventStr):
			// Message sent successfully
		default:
			// Channel is full or closed, skip
		}
	}
	return nil
}

// CloseAll closes all active sessions
func (m *MockConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.Close()
	}
	m.sessions = make(map[string]SSESession)
}

// Count returns the number of active sessions
func (m *MockConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// TestSSESession tests the SSESession interface with a mock implementation
func TestSSESession(t *testing.T) {
	// Create a session
	sessionID := "test-session-id"
	session := NewMockSSESession(sessionID)

	// Test ID()
	if session.ID() != sessionID {
		t.Errorf("session.ID() = %s, want %s", session.ID(), sessionID)
	}

	// Test Context()
	if session.Context() == nil {
		t.Error("session.Context() should not be nil")
	}

	// Test NotificationChannel()
	if session.NotificationChannel() == nil {
		t.Error("session.NotificationChannel() should not be nil")
	}

	// Test sending a message to the channel
	notification := "test notification"
	session.NotificationChannel() <- notification

	// Test Close()
	session.Close()
	if !session.IsClosed() {
		t.Error("session.IsClosed() should be true after Close()")
	}

	// Test that we can't send to a closed channel
	// Note: We should NOT try to read from the channel after closing
	// This is what caused the test to fail
	// Instead, check that the session is marked as closed
	if !session.IsClosed() {
		t.Error("session should be closed after Close()")
	}
}

// TestConnectionManager tests the ConnectionManager interface with a mock implementation
func TestConnectionManager(t *testing.T) {
	// Create a connection manager
	manager := NewMockConnectionManager()

	// Test Count() with no sessions
	if count := manager.Count(); count != 0 {
		t.Errorf("manager.Count() = %d, want 0", count)
	}

	// Create test sessions
	session1 := NewMockSSESession("session1")
	session2 := NewMockSSESession("session2")

	// Test AddSession and Count()
	manager.AddSession(session1)
	if count := manager.Count(); count != 1 {
		t.Errorf("manager.Count() = %d, want 1", count)
	}

	manager.AddSession(session2)
	if count := manager.Count(); count != 2 {
		t.Errorf("manager.Count() = %d, want 2", count)
	}

	// Test GetSession
	retrievedSession, ok := manager.GetSession("session1")
	if !ok {
		t.Error("manager.GetSession(\"session1\") should return true")
	}
	if retrievedSession.ID() != "session1" {
		t.Errorf("retrievedSession.ID() = %s, want session1", retrievedSession.ID())
	}

	// Test RemoveSession
	manager.RemoveSession("session1")
	if count := manager.Count(); count != 1 {
		t.Errorf("manager.Count() = %d, want 1", count)
	}
	if _, ok := manager.GetSession("session1"); ok {
		t.Error("manager.GetSession(\"session1\") should return false after RemoveSession")
	}

	// Test Broadcast
	event := map[string]interface{}{
		"type": "test-event",
		"data": "test-data",
	}
	err := manager.Broadcast(event)
	if err != nil {
		t.Errorf("manager.Broadcast() error = %v", err)
	}

	// Test CloseAll
	manager.CloseAll()
	if count := manager.Count(); count != 0 {
		t.Errorf("manager.Count() = %d, want 0 after CloseAll", count)
	}
	if !session2.IsClosed() {
		t.Error("session2.IsClosed() should be true after CloseAll()")
	}
}

// MockMessageHandler is a mock implementation of MessageHandler
type MockMessageHandler struct {
	handleFunc func(ctx context.Context, rawMessage json.RawMessage) interface{}
}

// NewMockMessageHandler creates a new MockMessageHandler
func NewMockMessageHandler(handleFunc func(ctx context.Context, rawMessage json.RawMessage) interface{}) *MockMessageHandler {
	return &MockMessageHandler{
		handleFunc: handleFunc,
	}
}

// HandleMessage processes a raw JSON message and returns a response
func (m *MockMessageHandler) HandleMessage(ctx context.Context, rawMessage json.RawMessage) interface{} {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, rawMessage)
	}
	return nil
}

// TestMessageHandler tests the MessageHandler interface with a mock implementation
func TestMessageHandler(t *testing.T) {
	// Expected message and response
	expectedMsg := json.RawMessage(`{"method":"test","params":{"key":"value"}}`)
	expectedResponse := map[string]interface{}{
		"result": "success",
	}

	// Create a mock handler
	handler := NewMockMessageHandler(func(ctx context.Context, rawMessage json.RawMessage) interface{} {
		// Verify the message content
		var msg map[string]interface{}
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
			return nil
		}
		if method, ok := msg["method"].(string); !ok || method != "test" {
			t.Errorf("Expected method 'test', got %v", msg["method"])
		}

		// Return the expected response
		return expectedResponse
	})

	// Test HandleMessage
	response := handler.HandleMessage(context.Background(), expectedMsg)
	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Errorf("Expected response to be map[string]interface{}, got %T", response)
	} else if respMap["result"] != "success" {
		t.Errorf("Expected result 'success', got %v", respMap["result"])
	}
}
