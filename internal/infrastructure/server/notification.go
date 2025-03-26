package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

// JSONRPCNotification represents a notification sent to clients via JSON-RPC.
type JSONRPCNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// NotificationChannel is a channel for sending notifications.
type NotificationChannel chan JSONRPCNotification

// MCPSession represents a connected client session.
type MCPSession struct {
	id        string
	userAgent string
	notifChan NotificationChannel
}

// NewMCPSession creates a new MCPSession.
func NewMCPSession(id, userAgent string, bufferSize int) *MCPSession {
	return &MCPSession{
		id:        id,
		userAgent: userAgent,
		notifChan: make(NotificationChannel, bufferSize),
	}
}

// ID returns the session ID.
func (s *MCPSession) ID() string {
	return s.id
}

// NotificationChannel returns the channel for sending notifications to this session.
func (s *MCPSession) NotificationChannel() NotificationChannel {
	return s.notifChan
}

// Close closes the notification channel.
func (s *MCPSession) Close() {
	close(s.notifChan)
}

// NotificationSender handles sending notifications to clients.
type NotificationSender struct {
	sessions       sync.Map
	jsonrpcVersion string
}

// NewNotificationSender creates a new NotificationSender.
func NewNotificationSender(jsonrpcVersion string) *NotificationSender {
	return &NotificationSender{
		jsonrpcVersion: jsonrpcVersion,
	}
}

// RegisterSession registers a session for notifications.
func (n *NotificationSender) RegisterSession(session *MCPSession) {
	n.sessions.Store(session.ID(), session)
}

// UnregisterSession unregisters a session.
func (n *NotificationSender) UnregisterSession(sessionID string) {
	if session, ok := n.sessions.LoadAndDelete(sessionID); ok {
		session.(*MCPSession).Close()
	}
}

// SendNotification sends a notification to a specific client.
func (n *NotificationSender) SendNotification(ctx context.Context, sessionID string, notification *domain.Notification) error {
	value, ok := n.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session := value.(*MCPSession)
	jsonRPC := JSONRPCNotification{
		JSONRPC: n.jsonrpcVersion,
		Method:  notification.Method,
		Params:  notification.Params,
	}

	select {
	case session.NotificationChannel() <- jsonRPC:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("notification channel for session %s is full or closed", sessionID)
	}
}

// BroadcastNotification sends a notification to all connected clients.
func (n *NotificationSender) BroadcastNotification(ctx context.Context, notification *domain.Notification) error {
	jsonRPC := JSONRPCNotification{
		JSONRPC: n.jsonrpcVersion,
		Method:  notification.Method,
		Params:  notification.Params,
	}

	var wg sync.WaitGroup
	var errsMu sync.Mutex
	var errs []error

	// Function to process a session
	processSession := func(key, value interface{}) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			session := value.(*MCPSession)

			// Try to send notification with timeout from context
			select {
			case session.NotificationChannel() <- jsonRPC:
				// Successfully sent
			case <-ctx.Done():
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("context cancelled for session %s: %w", session.ID(), ctx.Err()))
				errsMu.Unlock()
			default:
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("notification channel for session %s is full or closed", session.ID()))
				errsMu.Unlock()
			}
		}()
		return true
	}

	// Process all sessions
	n.sessions.Range(processSession)

	// Wait for all goroutines to complete
	wg.Wait()

	// Return first error if any occurred
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}
