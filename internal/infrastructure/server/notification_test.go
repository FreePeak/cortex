package server

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJsonrpcVersion = "2.0"

// Test NewMCPSession
func TestNewMCPSession(t *testing.T) {
	sessionID := "test-session-1"
	userAgent := "test-agent"
	bufferSize := 10

	session := NewMCPSession(sessionID, userAgent, bufferSize)

	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID())
	assert.Equal(t, userAgent, session.userAgent) // Accessing unexported field for test validation
	assert.NotNil(t, session.NotificationChannel())
	assert.Equal(t, bufferSize, cap(session.NotificationChannel()))
}

// Test MCPSession ID method
func TestMCPSession_ID(t *testing.T) {
	sessionID := "test-session-id"
	session := NewMCPSession(sessionID, "agent", 5)
	assert.Equal(t, sessionID, session.ID())
}

// Test MCPSession NotificationChannel method
func TestMCPSession_NotificationChannel(t *testing.T) {
	session := NewMCPSession("id", "agent", 5)
	assert.NotNil(t, session.NotificationChannel())
}

// Test MCPSession Close method
func TestMCPSession_Close(t *testing.T) {
	session := NewMCPSession("id", "agent", 5)
	ch := session.NotificationChannel()

	session.Close()

	// Check if channel is closed
	_, ok := <-ch
	assert.False(t, ok, "Channel should be closed")
}

// Test NewNotificationSender
func TestNewNotificationSender(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	assert.NotNil(t, sender)
	assert.Equal(t, testJsonrpcVersion, sender.jsonrpcVersion)
}

// Test RegisterSession and UnregisterSession
func TestNotificationSender_RegisterUnregisterSession(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	sessionID := "session-to-register"
	session := NewMCPSession(sessionID, "agent", 5)

	// Register
	sender.RegisterSession(session)
	_, ok := sender.sessions.Load(sessionID)
	assert.True(t, ok, "Session should be registered")

	// Unregister
	sender.UnregisterSession(sessionID)
	_, ok = sender.sessions.Load(sessionID)
	assert.False(t, ok, "Session should be unregistered")

	// Check if channel was closed on unregister
	_, chanOk := <-session.NotificationChannel()
	assert.False(t, chanOk, "Session channel should be closed upon unregistration")

	// Test unregistering non-existent session (should not panic)
	sender.UnregisterSession("non-existent-session")
}

// Test SendNotification - Success case
func TestNotificationSender_SendNotification_Success(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	sessionID := "target-session"
	session := NewMCPSession(sessionID, "agent", 1) // Buffer size 1
	sender.RegisterSession(session)
	defer sender.UnregisterSession(sessionID)

	notification := &domain.Notification{
		Method: "test/method",
		Params: map[string]interface{}{"key": "value"},
	}

	ctx := context.Background()
	err := sender.SendNotification(ctx, sessionID, notification)
	require.NoError(t, err)

	// Verify notification was received
	select {
	case receivedNotif := <-session.NotificationChannel():
		assert.Equal(t, testJsonrpcVersion, receivedNotif.JSONRPC)
		assert.Equal(t, notification.Method, receivedNotif.Method)
		assert.Equal(t, notification.Params, receivedNotif.Params)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive notification in time")
	}
}

// Test SendNotification - Session not found
func TestNotificationSender_SendNotification_SessionNotFound(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	notification := &domain.Notification{Method: "test"}

	ctx := context.Background()
	err := sender.SendNotification(ctx, "non-existent-session", notification)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session non-existent-session not found")
}

// Test SendNotification - Channel full or closed
func TestNotificationSender_SendNotification_ChannelFullOrClosed(t *testing.T) { // Renamed for clarity
	sender := NewNotificationSender(testJsonrpcVersion)
	sessionID := "full-channel-session"
	session := NewMCPSession(sessionID, "agent", 0) // Buffer size 0 - will block immediately
	sender.RegisterSession(session)
	// Do not start a receiver goroutine

	notification := &domain.Notification{Method: "test"}
	ctx := context.Background()

	sendAttempted := make(chan struct{}) // Channel to signal send attempt
	var wg sync.WaitGroup
	var err error
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Signal that we are about to attempt the send, potentially blocking
		close(sendAttempted)
		// This send will block because the channel is unbuffered and no one is reading.
		// It will error out when the channel is closed by UnregisterSession.
		err = sender.SendNotification(ctx, sessionID, notification)
	}()

	// Wait until the goroutine signals it's attempting the send
	<-sendAttempted

	// Give the send a very brief moment to potentially enter the blocking select case within SendNotification.
	// This helps ensure the send is actually blocked before we close the channel.
	// Still a small timing element, but more targeted than the previous fixed sleep.
	time.Sleep(5 * time.Millisecond)

	// Now unregister the session, which closes the channel, causing the blocked send to error out.
	sender.UnregisterSession(sessionID)

	wg.Wait() // Wait for the SendNotification goroutine to finish

	require.Error(t, err, "Expected an error when sending to a closed/full channel")
	// Check that the error message indicates the problem, potentially mentioning the session.
	// The exact wording might depend on SendNotification's internal error handling (e.g., context vs channel state).
	assert.Contains(t, err.Error(), sessionID, "Error message should ideally mention the session ID")
	// Example of a more specific check if the error is known:
	// assert.Contains(t, err.Error(), "channel closed", "Error message should indicate channel closure")
}

// Test SendNotification - Context cancelled
func TestNotificationSender_SendNotification_ContextCancelled(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	sessionID := "ctx-cancel-session"
	// Use buffer size 0 to make it easier to hit the context cancellation path
	session := NewMCPSession(sessionID, "agent", 0)
	sender.RegisterSession(session)
	defer sender.UnregisterSession(sessionID) // Ensure cleanup

	notification := &domain.Notification{Method: "test"}
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	err := sender.SendNotification(ctx, sessionID, notification)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// Test BroadcastNotification - Success case
func TestNotificationSender_BroadcastNotification_Success(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	session1 := NewMCPSession("s1", "agent1", 1)
	session2 := NewMCPSession("s2", "agent2", 1)
	sender.RegisterSession(session1)
	sender.RegisterSession(session2)
	defer sender.UnregisterSession("s1")
	defer sender.UnregisterSession("s2")

	notification := &domain.Notification{
		Method: "broadcast/test",
		Params: map[string]interface{}{"data": 123},
	}

	ctx := context.Background()
	err := sender.BroadcastNotification(ctx, notification)
	require.NoError(t, err)

	// Verify both sessions received the notification using WaitGroup
	var wg sync.WaitGroup
	wg.Add(2) // Expecting two notifications

	expectedNotif := JSONRPCNotification{
		JSONRPC: testJsonrpcVersion,
		Method:  notification.Method,
		Params:  notification.Params,
	}

	// Start listeners before broadcasting
	go func() {
		defer wg.Done()
		select {
		case received := <-session1.NotificationChannel():
			assert.Equal(t, expectedNotif, received, "Session 1 received incorrect notification")
		case <-time.After(200 * time.Millisecond): // Increased timeout for safety
			assert.Fail(t, "Timeout waiting for notification on session 1")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case received := <-session2.NotificationChannel():
			assert.Equal(t, expectedNotif, received, "Session 2 received incorrect notification")
		case <-time.After(200 * time.Millisecond): // Increased timeout for safety
			assert.Fail(t, "Timeout waiting for notification on session 2")
		}
	}()

	// Wait for both listeners to receive (or timeout)
	// Use a channel to wait for the WaitGroup to avoid blocking the main test goroutine indefinitely
	// if wg.Wait() hangs.
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// WaitGroup finished, assertions were checked in goroutines.
	case <-time.After(500 * time.Millisecond): // Overall timeout for the wait group
		t.Fatal("Test timed out waiting for broadcast notifications")
	}
}

// Test BroadcastNotification - One channel full
func TestNotificationSender_BroadcastNotification_OneChannelFull(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	sessionOK := NewMCPSession("sOK", "agentOK", 1)
	sessionFull := NewMCPSession("sFull", "agentFull", 0) // Buffer 0, will block
	sender.RegisterSession(sessionOK)
	sender.RegisterSession(sessionFull)
	defer sender.UnregisterSession("sOK")
	defer sender.UnregisterSession("sFull")

	notification := &domain.Notification{Method: "broadcast/full"}

	ctx := context.Background()
	err := sender.BroadcastNotification(ctx, notification)
	require.Error(t, err) // Expect an error because one channel is full
	assert.Contains(t, err.Error(), "notification channel for session sFull is full or closed")

	// Verify the OK session still received it
	select {
	case <-sessionOK.NotificationChannel():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Session sOK did not receive broadcast")
	}
}

// Test BroadcastNotification - Context cancelled during broadcast
func TestNotificationSender_BroadcastNotification_ContextCancelled(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)

	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Create some sessions
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("s%d", i)
		sessions := NewMCPSession(id, "agent", 1)
		sender.RegisterSession(sessions)
		defer sender.UnregisterSession(id)
	}

	notification := &domain.Notification{Method: "broadcast/cancel"}

	// Now try to broadcast with the already-cancelled context
	err := sender.BroadcastNotification(ctx, notification)

	// Verify we get a context.Canceled error
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// Test BroadcastNotification - No sessions
func TestNotificationSender_BroadcastNotification_NoSessions(t *testing.T) {
	sender := NewNotificationSender(testJsonrpcVersion)
	notification := &domain.Notification{Method: "broadcast/empty"}

	ctx := context.Background()
	err := sender.BroadcastNotification(ctx, notification)
	require.NoError(t, err) // Should not error if there are no sessions
}
