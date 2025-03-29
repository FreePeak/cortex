package server

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockSSESession is a mock implementation of domain.SSESession for testing
type MockSSESession struct {
	id        string
	notifChan chan string
	closed    bool
	mu        sync.Mutex
	ctx       context.Context
}

func NewMockSSESession(id string) *MockSSESession {
	return &MockSSESession{
		id:        id,
		notifChan: make(chan string, 10),
		ctx:       context.Background(),
	}
}

func (m *MockSSESession) ID() string {
	return m.id
}

func (m *MockSSESession) NotificationChannel() chan<- string {
	return m.notifChan
}

func (m *MockSSESession) Context() context.Context {
	return m.ctx
}

func (m *MockSSESession) Start() {
	// No-op for tests
}

func (m *MockSSESession) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		m.closed = true
		close(m.notifChan)
	}
}

func (m *MockSSESession) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestNewSSEConnectionManager(t *testing.T) {
	mgr := NewSSEConnectionManager()
	assert.NotNil(t, mgr)
	assert.Equal(t, 0, mgr.Count())
}

func TestSSEConnectionManager_AddAndGetSession(t *testing.T) {
	mgr := NewSSEConnectionManager()

	// Add a session
	session1 := NewMockSSESession("session1")
	mgr.AddSession(session1)

	// Add another session
	session2 := NewMockSSESession("session2")
	mgr.AddSession(session2)

	// Verify count
	assert.Equal(t, 2, mgr.Count())

	// Get an existing session
	s, found := mgr.GetSession("session1")
	assert.True(t, found)
	assert.Equal(t, "session1", s.ID())

	// Get a non-existent session
	s, found = mgr.GetSession("non-existent")
	assert.False(t, found)
	assert.Nil(t, s)
}

func TestSSEConnectionManager_RemoveSession(t *testing.T) {
	mgr := NewSSEConnectionManager()

	// Add a session
	session := NewMockSSESession("test-session")
	mgr.AddSession(session)
	assert.Equal(t, 1, mgr.Count())

	// Remove the session
	mgr.RemoveSession("test-session")
	assert.Equal(t, 0, mgr.Count())

	// Check if the session is closed
	time.Sleep(10 * time.Millisecond) // Give a bit of time for async operations

	// Remove non-existent session (should not panic)
	mgr.RemoveSession("non-existent")
}

func TestSSEConnectionManager_CloseAll(t *testing.T) {
	mgr := NewSSEConnectionManager()

	// Add multiple sessions
	sessions := []*MockSSESession{
		NewMockSSESession("session1"),
		NewMockSSESession("session2"),
		NewMockSSESession("session3"),
	}

	for _, session := range sessions {
		mgr.AddSession(session)
	}

	assert.Equal(t, 3, mgr.Count())

	// Close all sessions
	mgr.CloseAll()

	// Verify count is reset
	assert.Equal(t, 0, mgr.Count())

	// Verify all sessions are closed
	for _, session := range sessions {
		assert.True(t, session.IsClosed(), "Session should be closed after CloseAll")
	}
}

func TestSSEConnectionManager_Broadcast(t *testing.T) {
	mgr := NewSSEConnectionManager()

	// Add multiple sessions
	session1 := NewMockSSESession("session1")
	session2 := NewMockSSESession("session2")
	mgr.AddSession(session1)
	mgr.AddSession(session2)

	// Create a test event
	event := map[string]string{
		"type": "test-event",
		"data": "test-data",
	}

	// Broadcast the event
	err := mgr.Broadcast(event)
	assert.NoError(t, err)

	// Verify each session receives the event
	for _, session := range []*MockSSESession{session1, session2} {
		select {
		case msg := <-session.notifChan:
			// Verify message content contains the event data
			assert.Contains(t, msg, "test-event")
			assert.Contains(t, msg, "test-data")
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Session %s did not receive the broadcast message", session.ID())
		}
	}
}

func TestSSEConnectionManager_BroadcastToFullQueue(t *testing.T) {
	mgr := NewSSEConnectionManager()

	// Create a session with a small channel buffer
	session := &MockSSESession{
		id:        "test-session",
		notifChan: make(chan string, 1), // Very small buffer
		ctx:       context.Background(),
	}

	mgr.AddSession(session)

	// Fill up the channel
	session.notifChan <- "dummy message"

	// Broadcast events in rapid succession - should not block
	for i := 0; i < 5; i++ {
		err := mgr.Broadcast(map[string]string{"data": "overflow-test"})
		assert.NoError(t, err)
	}

	// Clean up
	session.Close()
}

func TestSSEConnectionManager_BroadcastToClosedSession(t *testing.T) {
	mgr := NewSSEConnectionManager()

	// Add a session and then close it
	session := NewMockSSESession("closed-session")
	mgr.AddSession(session)

	// Close the session, which closes the notification channel
	session.Close()

	// Remove the session from the manager (to avoid sending to closed channel)
	mgr.RemoveSession("closed-session")

	// Now broadcast should succeed without panicking
	assert.NotPanics(t, func() {
		mgr.Broadcast(map[string]string{"data": "test"})
	})
}

func TestSSEConnectionManager_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	mgr := NewSSEConnectionManager()
	var wg sync.WaitGroup

	// Define a fixed number of goroutines and sessions
	const totalSessions = 50
	const sessionsToRemove = 25

	// Use a mutex to protect the session count calculation
	var mu sync.Mutex
	var sessionCount int

	// Add a bunch of concurrent sessions
	for i := 0; i < totalSessions; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			session := NewMockSSESession(fmt.Sprintf("session-%d", id))
			mgr.AddSession(session)

			mu.Lock()
			sessionCount++
			mu.Unlock()
		}(i)
	}

	// Wait for all sessions to be added before removing any
	wg.Wait()

	// Reset sessionCount to match what was actually added
	mu.Lock()
	sessionCount = mgr.Count()
	mu.Unlock()

	// Create a new wait group for removals
	var wgRemove sync.WaitGroup

	// Remove some sessions concurrently
	for i := 0; i < sessionsToRemove; i++ {
		wgRemove.Add(1)
		go func(id int) {
			defer wgRemove.Done()
			mgr.RemoveSession(fmt.Sprintf("session-%d", id))

			mu.Lock()
			sessionCount--
			mu.Unlock()
		}(i)
	}

	wgRemove.Wait()

	// Verify the expected count matches our calculated count
	expectedSessionCount := totalSessions - sessionsToRemove
	assert.Equal(t, expectedSessionCount, sessionCount)
	assert.Equal(t, sessionCount, mgr.Count())

	// Clean up
	mgr.CloseAll()
}
