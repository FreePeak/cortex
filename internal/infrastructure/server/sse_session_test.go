package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResponseWriterFlusher implements http.ResponseWriter and http.Flusher
type mockResponseWriterFlusher struct {
	*httptest.ResponseRecorder
	flushed bool
}

func newMockResponseWriterFlusher() *mockResponseWriterFlusher {
	return &mockResponseWriterFlusher{
		ResponseRecorder: httptest.NewRecorder(),
		flushed:          false,
	}
}

func (m *mockResponseWriterFlusher) Flush() {
	m.flushed = true
}

// strictNonFlusherWriter implements http.ResponseWriter but *not* http.Flusher
type strictNonFlusherWriter struct{}

func (w *strictNonFlusherWriter) Header() http.Header {
	return http.Header{}
}

func (w *strictNonFlusherWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (w *strictNonFlusherWriter) WriteHeader(statusCode int) {
	// Do nothing
}

func TestNewSSESession(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		w := newMockResponseWriterFlusher()
		session, err := NewSSESession(w, "test-agent", 10)

		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.NotEmpty(t, session.ID())
	})

	t.Run("Error_NonFlusher", func(t *testing.T) {
		// Use our custom implementation that doesn't implement http.Flusher
		w := &strictNonFlusherWriter{}

		session, err := NewSSESession(w, "test-agent", 10)

		assert.Error(t, err)
		assert.Equal(t, ErrResponseWriterNotFlusher, err)
		assert.Nil(t, session)
	})
}

func TestSSESession_ID(t *testing.T) {
	w := newMockResponseWriterFlusher()
	session, err := NewSSESession(w, "test-agent", 10)
	require.NoError(t, err)

	id := session.ID()
	assert.NotEmpty(t, id)
}

func TestSSESession_NotificationChannel(t *testing.T) {
	w := newMockResponseWriterFlusher()
	session, err := NewSSESession(w, "test-agent", 10)
	require.NoError(t, err)

	ch := session.NotificationChannel()
	assert.NotNil(t, ch)
}

func TestSSESession_Close(t *testing.T) {
	w := newMockResponseWriterFlusher()
	session, err := NewSSESession(w, "test-agent", 10)
	require.NoError(t, err)

	// Verify that Close doesn't panic
	assert.NotPanics(t, func() {
		session.Close()
	})

	// Verify that the context is canceled
	ctx := session.Context()
	select {
	case <-ctx.Done():
		// Context should be canceled
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context was not canceled within the expected timeframe")
	}
}

func TestSSESession_Context(t *testing.T) {
	w := newMockResponseWriterFlusher()
	session, err := NewSSESession(w, "test-agent", 10)
	require.NoError(t, err)

	ctx := session.Context()
	assert.NotNil(t, ctx)

	// Context should not be canceled initially
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be canceled yet")
	default:
		// This is the expected path
	}

	// After closing the session, the context should be canceled
	session.Close()
	select {
	case <-ctx.Done():
		// Context should be canceled
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context was not canceled after session close")
	}
}

func TestSSESession_Start(t *testing.T) {
	w := newMockResponseWriterFlusher()
	session, err := NewSSESession(w, "test-agent", 10)
	require.NoError(t, err)

	// Start session in a goroutine
	go func() {
		session.Start()
	}()

	// Give it a moment to set headers
	time.Sleep(50 * time.Millisecond)

	// Verify headers are set correctly
	headers := w.ResponseRecorder.Header()
	assert.Equal(t, "text/event-stream", headers.Get("Content-Type"))
	assert.Equal(t, "no-cache", headers.Get("Cache-Control"))
	assert.Equal(t, "keep-alive", headers.Get("Connection"))
	assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
	assert.True(t, w.flushed, "The response should be flushed")

	// Send an event to the session
	testEvent := "event: test\ndata: test message\n\n"
	session.NotificationChannel() <- testEvent

	// Give it a moment to process the event
	time.Sleep(50 * time.Millisecond)

	// Verify the event was written to the response
	responseBody := w.ResponseRecorder.Body.String()
	assert.Contains(t, responseBody, "event: test")
	assert.Contains(t, responseBody, "data: test message")

	// Clean up
	session.Close()
}

func TestSSESession_StartWithNotification(t *testing.T) {
	w := newMockResponseWriterFlusher()
	session, err := NewSSESession(w, "test-agent", 10)
	require.NoError(t, err)

	// Start session in a goroutine
	go func() {
		session.Start()
	}()

	// Get the internal session to access notifChan
	sseSession, ok := session.(*sseSession2)
	require.True(t, ok)

	// Give it a moment to set headers
	time.Sleep(50 * time.Millisecond)

	// Send a notification through the notification channel
	notification := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  "test.method",
		Params:  map[string]interface{}{"key": "value"},
	}

	sseSession.notifChan <- notification

	// Give it a moment to process the notification
	time.Sleep(50 * time.Millisecond)

	// Verify the notification was converted to an SSE event and written to the response
	responseBody := w.ResponseRecorder.Body.String()
	assert.Contains(t, responseBody, "event: message")
	assert.Contains(t, responseBody, `"jsonrpc":"2.0"`)
	assert.Contains(t, responseBody, `"method":"test.method"`)

	// Clean up
	session.Close()
}
