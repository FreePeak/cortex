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
}

// NewSSEHandler creates a new SSE handler with the given configuration and dependencies.
func NewSSEHandler(
	config SSEHandlerConfig,
	messageHandler domain.MessageHandler,
	jsonrpcVersion string,
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

	// CORS headers for SSE
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Get user agent for session tracking
	userAgent := r.UserAgent()

	// Create a new session
	session, err := NewSSESession(w, userAgent, 100) // Buffer size of 100
	if err != nil {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Add session to connection manager
	s.connectionMgr.AddSession(session)

	// Start processing events (in the current goroutine)
	// This will block until the client disconnects
	go func() {
		// Remove the session when done
		defer s.connectionMgr.RemoveSession(session.ID())
		session.Start()
	}()

	// Block until the client disconnects or context is canceled
	<-session.Context().Done()
}

// handleMessage processes a JSON-RPC message request.
func (s *sseHandler) handleMessage(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests for messages
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONRPCError(w, nil, -32700, "Error reading request body", s.jsonrpcVersion)
		return
	}

	// Create a context with values from the request
	ctx := r.Context()

	// Call the message handler to process the message
	response := s.messageHandler.HandleMessage(ctx, body)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
