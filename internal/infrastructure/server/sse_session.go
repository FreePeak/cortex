package server

import (
	"context"
	"net/http"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/google/uuid"
)

// sse_session.go defines the implementation of the domain.SSESession interface.

// sseSession2 represents an active SSE connection and implements the domain.SSESession interface.
type sseSession2 struct {
	writer     http.ResponseWriter
	flusher    http.Flusher
	done       chan struct{}
	eventQueue chan string // Channel for queuing events
	id         string
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSSESession creates a new SSE session.
func NewSSESession(w http.ResponseWriter, userAgent string, bufferSize int) (domain.SSESession, error) {
	// Check if the ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrResponseWriterNotFlusher
	}

	// Create a context with cancellation for this session
	ctx, cancel := context.WithCancel(context.Background())

	// Generate a unique ID for this session
	sessionID := uuid.New().String()

	session := &sseSession2{
		writer:     w,
		flusher:    flusher,
		done:       make(chan struct{}),
		eventQueue: make(chan string, bufferSize),
		id:         sessionID,
		ctx:        ctx,
		cancel:     cancel,
	}

	return session, nil
}

// ID returns the session ID.
func (s *sseSession2) ID() string {
	return s.id
}

// NotificationChannel returns the channel for sending notifications.
func (s *sseSession2) NotificationChannel() chan<- string {
	return s.eventQueue
}

// Close closes the session.
func (s *sseSession2) Close() {
	s.cancel()
	close(s.done)
	// We intentionally don't close eventQueue here to avoid panic
	// when writing to a closed channel. It will be garbage collected
	// when the session object is no longer referenced.
}

// Start begins processing events for this session.
// This method should be called in a separate goroutine.
func (s *sseSession2) Start() {
	// Set headers for SSE
	s.writer.Header().Set("Content-Type", "text/event-stream")
	s.writer.Header().Set("Cache-Control", "no-cache")
	s.writer.Header().Set("Connection", "keep-alive")
	s.writer.Header().Set("Access-Control-Allow-Origin", "*")
	s.flusher.Flush()

	// Send an initial message
	s.writer.Write([]byte("event: connected\ndata: {\"id\":\"" + s.id + "\"}\n\n"))
	s.flusher.Flush()

	for {
		select {
		case <-s.ctx.Done():
			// Context cancelled, stop processing
			return
		case <-s.done:
			// Session closed, stop processing
			return
		case event := <-s.eventQueue:
			// Write the event directly to the response
			s.writer.Write([]byte(event))
			s.flusher.Flush()
		}
	}
}

// Context returns the session's context.
func (s *sseSession2) Context() context.Context {
	return s.ctx
}
