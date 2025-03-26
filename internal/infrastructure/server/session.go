package server

import "github.com/FreePeak/golang-mcp-server-sdk/internal/domain"

// ClientSession defines the interface for a client session.
type ClientSession interface {
	// SessionID returns the unique identifier for this session.
	SessionID() string

	// NotificationChannel returns a channel that can be used to send notifications to the client.
	NotificationChannel() chan<- domain.JSONRPCNotification
}

// mcpSession represents a client session.
type mcpSession struct {
	sessionID           string
	userAgent           string
	notificationChannel chan domain.JSONRPCNotification
}

func (s *mcpSession) SessionID() string {
	return s.sessionID
}

func (s *mcpSession) NotificationChannel() chan<- domain.JSONRPCNotification {
	return s.notificationChannel
}
