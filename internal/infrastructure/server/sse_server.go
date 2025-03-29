package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
	"github.com/google/uuid"
)

// sseSession represents an active SSE connection.
type sseSession struct {
	writer     http.ResponseWriter
	flusher    http.Flusher
	done       chan struct{}
	eventQueue chan string // Channel for queuing events
	id         string
	notifChan  NotificationChannel
	ctx        context.Context
	cancel     context.CancelFunc
}

// SessionID returns the session ID.
func (s *sseSession) ID() string {
	return s.id
}

// NotificationChannel returns the channel for sending notifications.
func (s *sseSession) NotificationChannel() NotificationChannel {
	return s.notifChan
}

// Close closes the notification channel and cancels the context.
func (s *sseSession) Close() {
	s.cancel()
	close(s.notifChan)
	close(s.done)
}

// SSEContextFunc is a function that takes an existing context and the current
// request and returns a potentially modified context based on the request
// content. This can be used to inject context values from headers, for example.
type SSEContextFunc func(ctx context.Context, r *http.Request) context.Context

// ConnectionPool manages active SSE sessions.
type ConnectionPool struct {
	mu       sync.RWMutex
	sessions map[string]*sseSession
}

// NewConnectionPool creates a new connection pool.
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		sessions: make(map[string]*sseSession),
	}
}

// Add adds a session to the pool.
func (p *ConnectionPool) Add(session *sseSession) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[session.id] = session
}

// Remove removes a session from the pool.
func (p *ConnectionPool) Remove(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sessions, sessionID)
}

// Get returns a session by ID.
func (p *ConnectionPool) Get(sessionID string) (*sseSession, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	session, ok := p.sessions[sessionID]
	return session, ok
}

// Broadcast sends an event to all active sessions.
func (p *ConnectionPool) Broadcast(event interface{}) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return
	}

	eventStr := fmt.Sprintf("event: message\ndata: %s\n\n", eventData)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, session := range p.sessions {
		select {
		case session.eventQueue <- eventStr:
			// Event queued successfully
		case <-session.done:
			// Session is closed
		default:
			// Queue is full
		}
	}
}

// CloseAll closes all active sessions.
func (p *ConnectionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, session := range p.sessions {
		session.Close()
	}

	// Clear the map
	p.sessions = make(map[string]*sseSession)
}

// Count returns the number of active sessions.
func (p *ConnectionPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

// SSEServer implements a Server-Sent Events (SSE) based server.
// It provides real-time communication capabilities over HTTP using the SSE protocol.
type SSEServer struct {
	notifier        *NotificationSender
	baseURL         string
	basePath        string
	messageEndpoint string
	sseEndpoint     string
	connectionPool  *ConnectionPool
	srv             *http.Server
	contextFunc     SSEContextFunc
	mcpHandler      func(ctx context.Context, rawMessage json.RawMessage) interface{}
	logger          *logging.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

// SSEOption defines a function type for configuring SSEServer
type SSEOption func(*SSEServer)

// WithLogger sets the logger for the SSE server
func WithLogger(logger *logging.Logger) SSEOption {
	return func(s *SSEServer) {
		s.logger = logger
	}
}

// WithBaseURL sets the base URL for the SSE server
func WithBaseURL(baseURL string) SSEOption {
	return func(s *SSEServer) {
		if baseURL != "" {
			u, err := url.Parse(baseURL)
			if err != nil {
				return
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				return
			}
			// Check if the host is empty or only contains a port
			if u.Host == "" || strings.HasPrefix(u.Host, ":") {
				return
			}
			if len(u.Query()) > 0 {
				return
			}
		}
		s.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithBasePath sets the base path for the SSE server
func WithBasePath(basePath string) SSEOption {
	return func(s *SSEServer) {
		// Ensure the path starts with / and doesn't end with /
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		s.basePath = strings.TrimSuffix(basePath, "/")
	}
}

// WithMessageEndpoint sets the message endpoint path
func WithMessageEndpoint(endpoint string) SSEOption {
	return func(s *SSEServer) {
		s.messageEndpoint = endpoint
	}
}

// WithSSEEndpoint sets the SSE endpoint path
func WithSSEEndpoint(endpoint string) SSEOption {
	return func(s *SSEServer) {
		s.sseEndpoint = endpoint
	}
}

// WithHTTPServer sets the HTTP server instance
func WithHTTPServer(srv *http.Server) SSEOption {
	return func(s *SSEServer) {
		s.srv = srv
	}
}

// WithSSEContextFunc sets a function that will be called to customize the context
// to the server using the incoming request.
func WithSSEContextFunc(fn SSEContextFunc) SSEOption {
	return func(s *SSEServer) {
		s.contextFunc = fn
	}
}

// NewSSEServer creates a new SSE server instance with the given notification sender and options.
func NewSSEServer(notifier *NotificationSender, mcpHandler func(ctx context.Context, rawMessage json.RawMessage) interface{}, opts ...SSEOption) *SSEServer {
	ctx, cancel := context.WithCancel(context.Background())

	// Create default logger
	defaultLogger, err := logging.New(logging.Config{
		Level:       logging.InfoLevel,
		Development: true,
		OutputPaths: []string{"stdout"},
		InitialFields: logging.Fields{
			"component": "sse-server",
		},
	})
	if err != nil {
		// Fallback to a simple default logger if we can't create the structured one
		defaultLogger = logging.Default()
	}

	s := &SSEServer{
		notifier:        notifier,
		sseEndpoint:     "/sse",
		messageEndpoint: "/message",
		mcpHandler:      mcpHandler,
		connectionPool:  NewConnectionPool(),
		logger:          defaultLogger,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Apply all options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// NewTestServer creates a test server for testing purposes
func NewTestServer(notifier *NotificationSender, mcpHandler func(ctx context.Context, rawMessage json.RawMessage) interface{}, opts ...SSEOption) *httptest.Server {
	sseServer := NewSSEServer(notifier, mcpHandler)
	for _, opt := range opts {
		opt(sseServer)
	}

	testServer := httptest.NewServer(sseServer)
	sseServer.baseURL = testServer.URL
	return testServer
}

// Start begins serving SSE connections on the specified address.
// It sets up HTTP handlers for SSE and message endpoints.
func (s *SSEServer) Start(addr string) error {
	s.srv = &http.Server{
		Addr:    addr,
		Handler: s,
	}

	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the SSE server, closing all active sessions
// and shutting down the HTTP server.
func (s *SSEServer) Shutdown(ctx context.Context) error {
	// Cancel the server context first to stop accepting new connections
	s.cancel()

	// Close all active sessions
	s.connectionPool.CloseAll()

	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

// handleSSE handles incoming SSE connection requests.
// It sets up appropriate headers and creates a new session for the client.
func (s *SSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Create a context for this session that is a child of the server context
	// and can be canceled when the session ends
	sessionCtx, sessionCancel := context.WithCancel(s.ctx)

	session := &sseSession{
		writer:     w,
		flusher:    flusher,
		done:       make(chan struct{}),
		eventQueue: make(chan string, 100), // Buffer for events
		id:         sessionID,
		notifChan:  make(NotificationChannel, 100),
		ctx:        sessionCtx,
		cancel:     sessionCancel,
	}

	// Add the session to the connection pool
	s.connectionPool.Add(session)
	defer s.connectionPool.Remove(sessionID)

	s.notifier.RegisterSession(&MCPSession{
		id:        sessionID,
		userAgent: r.UserAgent(),
		notifChan: session.notifChan,
	})
	defer s.notifier.UnregisterSession(sessionID)

	// Start notification handler for this session
	go func() {
		for {
			select {
			case notification := <-session.notifChan:
				eventData, err := json.Marshal(notification)
				if err == nil {
					select {
					case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
						// Event queued successfully
					case <-session.done:
						return
					case <-session.ctx.Done():
						return
					}
				}
			case <-session.done:
				return
			case <-session.ctx.Done():
				return
			case <-r.Context().Done():
				return
			}
		}
	}()

	messageEndpoint := fmt.Sprintf("%s?sessionId=%s", s.CompleteMessageEndpoint(), sessionID)

	// Send the initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"sessionId\": \"%s\"}\n\n", sessionID)
	flusher.Flush()

	// Send the endpoint event
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageEndpoint)
	flusher.Flush()

	// Main event loop - this runs in the HTTP handler goroutine
	for {
		select {
		case event := <-session.eventQueue:
			// Write the event to the response
			fmt.Fprint(w, event)
			flusher.Flush()
		case <-r.Context().Done():
			sessionCancel()
			close(session.done)
			return
		case <-session.ctx.Done():
			close(session.done)
			return
		}
	}
}

// handleMessage processes incoming JSON-RPC messages from clients and sends responses
// back through both the SSE connection and HTTP response.
func (s *SSEServer) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONRPCError(w, nil, -32600, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		s.writeJSONRPCError(w, nil, -32602, "Missing sessionId")
		return
	}

	session, ok := s.connectionPool.Get(sessionID)
	if !ok {
		s.writeJSONRPCError(w, nil, -32602, "Invalid session ID")
		return
	}

	// Create context for the message handler
	ctx := r.Context()
	if s.contextFunc != nil {
		ctx = s.contextFunc(ctx, r)
	}

	// Parse message as raw JSON
	var rawMessage json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawMessage); err != nil {
		s.writeJSONRPCError(w, nil, -32700, "Parse error")
		return
	}

	// Process message through MCP handler
	response := s.mcpHandler(ctx, rawMessage)

	// Only send response if there is one (not for notifications)
	if response != nil {
		eventData, _ := json.Marshal(response)

		// Queue the event for sending via SSE
		select {
		case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
			// Event queued successfully
		case <-session.done:
			// Session is closed, don't try to queue
		case <-session.ctx.Done():
			// Session context was canceled
		default:
			// Queue is full, could log this
		}

		// Send HTTP response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	} else {
		// For notifications, just send 200 OK with no body
		w.WriteHeader(http.StatusOK)
	}
}

// writeJSONRPCError writes a JSON-RPC error response with the given error details.
func (s *SSEServer) writeJSONRPCError(
	w http.ResponseWriter,
	id interface{},
	code int,
	message string,
) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(response)
}

// SendEventToSession sends an event to a specific SSE session identified by sessionID.
// Returns an error if the session is not found or closed.
func (s *SSEServer) SendEventToSession(
	sessionID string,
	event interface{},
) error {
	session, ok := s.connectionPool.Get(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Queue the event for sending via SSE
	select {
	case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
		return nil
	case <-session.done:
		return fmt.Errorf("session closed")
	case <-session.ctx.Done():
		return fmt.Errorf("session context canceled")
	default:
		return fmt.Errorf("event queue full")
	}
}

// BroadcastEvent sends an event to all active SSE sessions.
func (s *SSEServer) BroadcastEvent(event interface{}) {
	s.connectionPool.Broadcast(event)
}

func (s *SSEServer) GetUrlPath(input string) (string, error) {
	parse, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %s: %w", input, err)
	}
	return parse.Path, nil
}

func (s *SSEServer) CompleteSseEndpoint() string {
	return s.baseURL + s.basePath + s.sseEndpoint
}

func (s *SSEServer) CompleteSsePath() string {
	path, err := s.GetUrlPath(s.CompleteSseEndpoint())
	if err != nil {
		return s.basePath + s.sseEndpoint
	}
	return path
}

func (s *SSEServer) CompleteMessageEndpoint() string {
	return s.baseURL + s.basePath + s.messageEndpoint
}

func (s *SSEServer) CompleteMessagePath() string {
	path, err := s.GetUrlPath(s.CompleteMessageEndpoint())
	if err != nil {
		return s.basePath + s.messageEndpoint
	}
	return path
}

// ServeHTTP implements the http.Handler interface.
func (s *SSEServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// Use exact path matching rather than Contains
	ssePath := s.CompleteSsePath()
	if ssePath != "" && path == ssePath {
		s.handleSSE(w, r)
		return
	}
	messagePath := s.CompleteMessagePath()
	if messagePath != "" && path == messagePath {
		s.handleMessage(w, r)
		return
	}

	http.NotFound(w, r)
}
