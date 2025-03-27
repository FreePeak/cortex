package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/google/uuid"
)

// SSEHandlerConfig contains configuration options for the SSE handler.
type SSEHandlerConfig struct {
	BaseURL         string
	BasePath        string
	MessageEndpoint string
	SSEEndpoint     string
}

// sseHandler implements the domain.SSEHandler interface.
type sseHandler struct {
	config         SSEHandlerConfig
	connectionMgr  domain.ConnectionManager
	messageHandler domain.MessageHandler
	httpServer     *http.Server
	jsonrpcVersion string
	notifier       *NotificationSender
}

// NewSSEHandler creates a new SSE handler with the given configuration and dependencies.
func NewSSEHandler(
	config SSEHandlerConfig,
	messageHandler domain.MessageHandler,
	jsonrpcVersion string,
	notifier *NotificationSender,
) domain.SSEHandler {
	// Set default endpoints if not provided
	if config.MessageEndpoint == "" {
		config.MessageEndpoint = "/message"
	}
	if config.SSEEndpoint == "" {
		config.SSEEndpoint = "/sse"
	}

	// Ensure paths start with / and don't end with /
	if !strings.HasPrefix(config.BasePath, "/") && config.BasePath != "" {
		config.BasePath = "/" + config.BasePath
	}
	config.BasePath = strings.TrimSuffix(config.BasePath, "/")

	return &sseHandler{
		config:         config,
		connectionMgr:  NewSSEConnectionManager(),
		messageHandler: messageHandler,
		jsonrpcVersion: jsonrpcVersion,
		notifier:       notifier,
	}
}

// ServeHTTP handles HTTP requests for SSE events.
func (s *sseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Normalize the request path for comparison
	path := r.URL.Path

	// Check if this is an SSE request
	if path == s.completeSSEPath() {
		s.handleSSE(w, r)
		return
	}

	// Check if this is a message request
	if path == s.completeMessagePath() {
		s.handleMessage(w, r)
		return
	}

	// Return 404 for other paths
	http.NotFound(w, r)
}

// Start starts the SSE server.
func (s *sseHandler) Start(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/", s) // Handle all requests through the ServeHTTP method

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting SSE server on %s", addr)
	log.Printf("SSE endpoint: %s", s.completeSSEPath())
	log.Printf("Message endpoint: %s", s.completeMessagePath())

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the SSE server.
func (s *sseHandler) Shutdown(ctx context.Context) error {
	// Close all active connections
	s.connectionMgr.CloseAll()

	// Shutdown the HTTP server if it exists
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// BroadcastEvent sends an event to all connected clients.
func (s *sseHandler) BroadcastEvent(event interface{}) error {
	return s.connectionMgr.Broadcast(event)
}

// SendEventToSession sends an event to a specific client session.
func (s *sseHandler) SendEventToSession(sessionID string, event interface{}) error {
	session, ok := s.connectionMgr.GetSession(sessionID)
	if !ok {
		return ErrSessionNotFound
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	eventStr := fmt.Sprintf("event: message\ndata: %s\n\n", eventData)

	select {
	case session.NotificationChannel() <- eventStr:
		return nil
	default:
		return ErrChannelFull
	}
}

// handleSSE processes a new SSE connection.
func (s *sseHandler) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Only support GET method for SSE
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if the ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Get or generate a session ID
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Get user agent for session tracking
	userAgent := r.UserAgent()

	// Create a context that cancels when the client disconnects
	// This is important for the original connection flow
	sessionCtx, sessionCancel := context.WithCancel(r.Context())

	// Create the event queue and notification channel
	eventQueue := make(chan string, 100)
	notifChan := make(NotificationChannel, 100)

	// Create a new session for the SSE connection
	session := &sseSession2{
		writer:     w,
		flusher:    flusher,
		done:       make(chan struct{}),
		eventQueue: eventQueue,
		id:         sessionID,
		notifChan:  notifChan,
		ctx:        sessionCtx,
		cancel:     sessionCancel,
	}

	// Add session to connection manager
	s.connectionMgr.AddSession(session)
	defer s.connectionMgr.RemoveSession(session.ID())

	// Register the session with the notification sender
	if s.notifier != nil {
		mcpSession := &MCPSession{
			id:        sessionID,
			userAgent: userAgent,
			notifChan: notifChan,
		}
		s.notifier.RegisterSession(mcpSession)
		defer s.notifier.UnregisterSession(sessionID)
	}

	// Create the message endpoint URL with session ID
	messageEndpoint := fmt.Sprintf("%s?sessionId=%s", s.completeMessagePath(), sessionID)

	// Send the initial connected event - IMPORTANT: this exact format is required
	fmt.Fprintf(w, "event: connected\ndata: {\"sessionId\": \"%s\"}\n\n", sessionID)
	flusher.Flush()

	// Send the endpoint event - IMPORTANT: this exact format is required
	fmt.Fprintf(w, "event: endpoint\ndata: \"%s\"\n\n", messageEndpoint)
	flusher.Flush()

	// Process events in the main goroutine
	// This matches the original SSE server's connection handling
	for {
		select {
		case event := <-eventQueue:
			// Write the event to the response
			fmt.Fprint(w, event)
			flusher.Flush()
		case notification := <-notifChan:
			// Convert notification to SSE event format
			eventData, err := json.Marshal(notification)
			if err == nil {
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", eventData)
				flusher.Flush()
			}
		case <-r.Context().Done():
			// Request context done (client disconnected)
			sessionCancel()
			close(session.done)
			return
		case <-sessionCtx.Done():
			// Session context done
			close(session.done)
			return
		}
	}
}

// handleMessage processes a JSON-RPC message request.
func (s *sseHandler) handleMessage(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests for messages
	if r.Method != http.MethodPost {
		writeJSONRPCError(w, nil, -32600, "Method not allowed", s.jsonrpcVersion)
		return
	}

	// CORS headers for message endpoint
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract sessionId from query parameters - this is critical
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		writeJSONRPCError(w, nil, -32602, "Missing sessionId parameter", s.jsonrpcVersion)
		return
	}

	// Verify session exists
	session, ok := s.connectionMgr.GetSession(sessionID)
	if !ok {
		writeJSONRPCError(w, nil, -32602, "Invalid session ID", s.jsonrpcVersion)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONRPCError(w, nil, -32700, "Error reading request body", s.jsonrpcVersion)
		return
	}

	// Create a context with values from the request
	ctx := r.Context()

	// Parse JSON-RPC request for logging and ID extraction
	var jsonRPCRequest map[string]interface{}
	if err := json.Unmarshal(body, &jsonRPCRequest); err == nil {
		// Log tool list requests
		if method, ok := jsonRPCRequest["method"].(string); ok && method == "tools/list" {
			log.Printf("Session %s requested tools/list", sessionID)
		}
	}

	// Process the message
	response := s.messageHandler.HandleMessage(ctx, body)

	// Only proceed if we have a response
	if response != nil {
		responseBytes, err := json.Marshal(response)
		if err != nil {
			writeJSONRPCError(w, nil, -32603, "Error marshalling response", s.jsonrpcVersion)
			return
		}

		// Get the request ID to determine if it's a request or notification
		id, hasID := jsonRPCRequest["id"]

		if hasID && id != nil {
			// It's a request (has an ID) - send both via SSE and HTTP

			// Send via SSE channel first
			select {
			case session.NotificationChannel() <- fmt.Sprintf("event: message\ndata: %s\n\n", responseBytes):
				// Event sent successfully
			default:
				log.Printf("Warning: Could not queue SSE event for session %s", sessionID)
			}

			// Also send HTTP response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBytes)
		} else {
			// It's a notification (no ID) - just send 202 Accepted
			w.WriteHeader(http.StatusAccepted)
		}
	} else {
		// No response (empty result)
		w.WriteHeader(http.StatusAccepted)
	}
}

// Helper function to get the request ID from a JSON-RPC request
func getRequestID(request map[string]interface{}) (interface{}, bool) {
	if request == nil {
		return nil, false
	}

	id, hasID := request["id"]
	return id, hasID
}

// Helper methods for path handling

func (s *sseHandler) completeSSEPath() string {
	return s.config.BasePath + s.config.SSEEndpoint
}

func (s *sseHandler) completeMessagePath() string {
	return s.config.BasePath + s.config.MessageEndpoint
}

// writeJSONRPCError writes a JSON-RPC error response to the ResponseWriter.
func writeJSONRPCError(
	w http.ResponseWriter,
	id interface{},
	code int,
	message string,
	jsonrpcVersion string,
) {
	w.Header().Set("Content-Type", "application/json")
	errResp := domain.CreateErrorResponse(jsonrpcVersion, id, code, message)
	json.NewEncoder(w).Encode(errResp)
}
