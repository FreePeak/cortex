package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/FreePeak/cortex/pkg/integration/pocketbase"
	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

// Since this is just an example, we're mocking PocketBase
// In a real application, you would import the actual PocketBase library
// import "github.com/pocketbase/pocketbase"
// import "github.com/pocketbase/pocketbase/core"

// MockPocketBase is a simple mock for PocketBase to demonstrate integration
type MockPocketBase struct {
	routes map[string]interface{}
	server *http.Server
}

func NewMockPocketBase() *MockPocketBase {
	return &MockPocketBase{
		routes: make(map[string]interface{}),
	}
}

// RegisterRoute simulates registering a route in PocketBase
func (m *MockPocketBase) RegisterRoute(path string, handler interface{}) {
	m.routes[path] = handler
	log.Printf("Registered route: %s", path)
}

// Start simulates starting a PocketBase server
func (m *MockPocketBase) Start(port int) error {
	log.Println("MockPocketBase server started")
	log.Println("Registered routes:")
	for path := range m.routes {
		log.Printf("  - %s", path)
	}

	// Create a unified handler that routes requests correctly
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For SSE requests we need to be extremely careful with headers
		if strings.HasSuffix(r.URL.Path, "/sse") {
			log.Printf("DEBUG: Incoming SSE request headers: %v", r.Header)
			log.Printf("DEBUG: Current response headers before routing: %v", w.Header())

			// For GET requests to SSE endpoint, immediately set SSE headers
			if r.Method == http.MethodGet {
				// Ensure no conflicting headers are present
				for k := range w.Header() {
					if k != "" && k != "Content-Type" {
						log.Printf("DEBUG: Clearing potential conflicting header: %s", k)
						w.Header().Del(k)
					}
				}

				// Set the Content-Type immediately and log it
				w.Header().Set("Content-Type", "text/event-stream")
				log.Printf("DEBUG: Pre-set Content-Type to text/event-stream")
			}
		}

		// First, check if the request is for the home page
		if r.URL.Path == "/" {
			fmt.Fprintf(w, "Mock PocketBase Server with Cortex MCP integration. Access MCP at /api/mcp")
			return
		}

		// Authorization endpoint for testing OAuth tokens
		if r.URL.Path == "/auth/token" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"access_token":"mock-token","token_type":"bearer","expires_in":3600,"scope":"cortex:tool:read cortex:tool:execute:echo"}`)
			return
		}

		// Special debug endpoint to check all registered routes
		if r.URL.Path == "/debug/routes" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "{\n  \"routes\": [\n")
			routes := make([]string, 0, len(m.routes))
			for route := range m.routes {
				routes = append(routes, route)
			}
			for i, route := range routes {
				if i < len(routes)-1 {
					fmt.Fprintf(w, "    \"%s\",\n", route)
				} else {
					fmt.Fprintf(w, "    \"%s\"\n", route)
				}
			}
			fmt.Fprintf(w, "  ],\n")
			fmt.Fprintf(w, "  \"request_path\": \"%s\"\n", r.URL.Path)
			fmt.Fprintf(w, "}\n")
			return
		}

		// Next, try to route to the registered handlers
		for path, handler := range m.routes {
			// Remove trailing /* for matching
			basePath := strings.TrimSuffix(path, "/*")

			// Check if the request path starts with this path
			if strings.HasPrefix(r.URL.Path, basePath) {
				if h, ok := handler.(http.Handler); ok {
					log.Printf("Routing request %s to handler for %s", r.URL.Path, basePath)

					// For SSE endpoints, we skip all middleware and route directly
					// This is crucial as SSE is very sensitive to headers
					if strings.HasSuffix(r.URL.Path, "/sse") {
						log.Printf("DEBUG: Direct routing to SSE handler, bypassing middleware")

						// Simply pass through to handler - our custom SSE wrapper will set headers
						h.ServeHTTP(w, r)
						return
					}

					h.ServeHTTP(w, r)
					return
				}
			}
		}

		// If no route matched, return 404
		http.NotFound(w, r)
	})

	// Create and start the server
	addr := fmt.Sprintf(":%d", port)
	m.server = &http.Server{
		Addr: addr,
		// CRITICAL: We bypass the SSE header preservation middleware as we're handling
		// this explicitly in our custom SSE handler now
		Handler: rootHandler,
	}

	log.Printf("Starting HTTP server on %s", addr)
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// This would be the main entry point for PocketBase plugins
type MockServeEvent struct{}

// OnBeforeServe simulates PocketBase's OnBeforeServe event
func (m *MockPocketBase) OnBeforeServe() *MockHook {
	return &MockHook{}
}

// MockHook simulates PocketBase's Hook system
type MockHook struct {
	handlers []func(*MockServeEvent) error
}

// Add simulates adding a handler to a PocketBase hook
func (h *MockHook) Add(handler func(*MockServeEvent) error) {
	h.handlers = append(h.handlers, handler)
}

// Trigger simulates triggering a PocketBase hook
func (h *MockHook) Trigger() error {
	for _, handler := range h.handlers {
		if err := handler(&MockServeEvent{}); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown simulates shutting down a PocketBase server
func (m *MockPocketBase) Shutdown(ctx context.Context) error {
	if m.server != nil {
		return m.server.Shutdown(ctx)
	}
	return nil
}

// Create a wrapper for the ResponseWriter specialized for SSE
type SSEResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	headersSent bool
	logger      *log.Logger
}

func NewSSEResponseWriter(w http.ResponseWriter) *SSEResponseWriter {
	return &SSEResponseWriter{
		ResponseWriter: w,
		logger:         log.New(os.Stderr, "[sse-writer] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// Header overrides the underlying Header method to ensure we never lose the SSE headers
func (w *SSEResponseWriter) Header() http.Header {
	// If headers have not been explicitly set yet, ensure SSE headers
	if !w.headersSent {
		// Always set/enforce these critical SSE headers
		h := w.ResponseWriter.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("Connection", "keep-alive")
		h.Set("Access-Control-Allow-Origin", "*")
		w.headersSent = true
		w.logger.Printf("Enforced SSE headers, Content-Type: %s", h.Get("Content-Type"))
	}
	return w.ResponseWriter.Header()
}

func (w *SSEResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		// Force SSE headers before writing the header
		w.Header()

		// Log what we're sending
		w.logger.Printf("WriteHeader with status %d, Content-Type: %s",
			statusCode, w.ResponseWriter.Header().Get("Content-Type"))

		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *SSEResponseWriter) Write(b []byte) (int, error) {
	// Force headers before first write
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// For debugging
	if len(b) < 100 {
		w.logger.Printf("Writing %d bytes: %s", len(b), string(b))
	} else {
		w.logger.Printf("Writing %d bytes", len(b))
	}

	return w.ResponseWriter.Write(b)
}

// This wrapper ensures we don't modify Content-Type once it's set
func (m *MockPocketBase) preserveSSEHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sse") {
			log.Printf("Using SSE ResponseWriter wrapper for: %s", r.URL.Path)

			// Important: For SSE endpoints, we need to ensure the Content-Type is set right
			// and preserved throughout the request handling
			if r.Method == http.MethodGet {
				// Pre-emptively set SSE headers to ensure they're preserved
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			sseWriter := NewSSEResponseWriter(w)
			next.ServeHTTP(sseWriter, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// MockTokenValidator simulates validating OAuth 2.1 tokens
type MockTokenValidator struct{}

// ValidateToken validates a mock token for testing
func (v *MockTokenValidator) ValidateToken(ctx context.Context, token string) (*server.TokenClaims, error) {
	// In a real application, this would validate the token with an auth server
	// For this example, we'll accept "mock-token" as valid
	if token == "mock-token" {
		return &server.TokenClaims{
			Subject:   "user123",
			Issuer:    "example-issuer",
			Audience:  []string{"cortex-api"},
			ExpiresAt: time.Now().Add(time.Hour),
			IssuedAt:  time.Now(),
			Scopes:    []string{"cortex:tool:read", "cortex:tool:execute:echo"},
			Claims:    map[string]interface{}{},
		}, nil
	}
	return nil, server.ErrInvalidToken
}

func main() {
	// Process command line flags
	var port int
	flag.IntVar(&port, "port", 8090, "Port to run the server on")
	flag.Parse()

	// Setup logging
	logger := log.New(os.Stderr, "[cortex] ", log.LstdFlags|log.Lmicroseconds)

	// Create a new Cortex plugin
	cortexPlugin := pocketbase.NewCortexPlugin(
		pocketbase.WithName("Cortex PocketBase Integration"),
		pocketbase.WithVersion("1.0.0"),
		pocketbase.WithLogger(logger),
		pocketbase.WithBasePath("/api/mcp"),
		pocketbase.WithPort(port),
	)

	// Setup OAuth 2.1 support
	// In a real application, you would use a proper token validator
	tokenValidator := &MockTokenValidator{}

	// Create OAuth middleware
	oauthMiddleware := server.NewOAuthMiddleware(tokenValidator)

	// Configure OAuth settings
	oauthConfig := &server.OAuthConfig{
		Issuer:            "example-issuer",
		Audience:          []string{"cortex-api"},
		TokenLookupScheme: "header,query",
		TokenHeaderName:   "Authorization",
		TokenQueryParam:   "access_token",
	}

	// Add OAuth to the plugin
	cortexPlugin.WithOAuth(oauthMiddleware).WithOAuthConfig(oauthConfig)

	// Add a tool
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	if err := cortexPlugin.AddTool(echoTool, handleEcho); err != nil {
		logger.Fatalf("Failed to add echo tool: %v", err)
	}

	// Add a weather tool
	weatherTool := tools.NewTool("weather",
		tools.WithDescription("Gets the weather for a location"),
		tools.WithString("location",
			tools.Description("The location to get weather for"),
			tools.Required(),
		),
	)

	if err := cortexPlugin.AddTool(weatherTool, handleWeather); err != nil {
		logger.Fatalf("Failed to add weather tool: %v", err)
	}

	// Create a mock PocketBase app
	pb := NewMockPocketBase()

	// Register the plugin with PocketBase
	if err := cortexPlugin.RegisterWithPocketBase(pb); err != nil {
		logger.Fatalf("Failed to register plugin: %v", err)
	}

	// Start the PocketBase app
	if err := pb.Start(port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}

	// Set up file serving for static files in the current directory
	// This allows us to serve the test client HTML
	workingDir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Failed to get working directory: %v", err)
	}

	http.Handle("/test/", http.StripPrefix("/test/", http.FileServer(http.Dir(workingDir))))

	// Create a data directory if it doesn't exist
	dataDir := filepath.Join(workingDir, "pb_data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			logger.Fatalf("Failed to create data directory: %v", err)
		}
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-stop

	logger.Println("Shutting down server...")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := pb.Shutdown(ctx); err != nil {
		logger.Fatalf("Error during shutdown: %v", err)
	}

	logger.Println("Server stopped")
}

// handleEcho is a tool handler that echoes back the input message
func handleEcho(ctx context.Context, request pocketbase.ToolCallRequest) (interface{}, error) {
	message, ok := request.Parameters["message"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid message parameter")
	}

	// Return the message
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}, nil
}

// handleWeather is a tool handler that simulates getting weather for a location
func handleWeather(ctx context.Context, request pocketbase.ToolCallRequest) (interface{}, error) {
	location, ok := request.Parameters["location"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid location parameter")
	}

	// Simulate a weather API call
	// In a real application, this would call a weather API
	weather := fmt.Sprintf("The weather in %s is sunny and 72°F", location)

	// Return the weather
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": weather,
			},
		},
	}, nil
}
