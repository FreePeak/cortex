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

func main() {
	// Parse command line flags
	var dataDir string
	var serverPort int
	flag.StringVar(&dataDir, "data", "./pb_data", "PocketBase data directory")
	flag.IntVar(&serverPort, "port", 8080, "Server port")
	flag.Parse()

	// Ensure the data directory exists
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v", err)
	}
	log.Printf("Using data directory: %s", absDataDir)

	// Create a logger with more context
	logger := log.New(os.Stderr, "[cortex] ", log.LstdFlags|log.Lmicroseconds)

	// Create a new PocketBase app
	app := NewMockPocketBase()

	// Initialize the Cortex plugin with custom options
	plugin := pocketbase.NewCortexPlugin(
		pocketbase.WithName("PocketBase MCP Server"),
		pocketbase.WithVersion("1.0.0"),
		pocketbase.WithBasePath("/api/mcp"),
		pocketbase.WithLogger(logger),
		pocketbase.WithPort(serverPort),
	)

	// Add an echo tool
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)
	plugin.AddTool(echoTool, handleEcho)

	// Add a weather tool
	weatherTool := tools.NewTool("weather",
		tools.WithDescription("Gets the weather for a location"),
		tools.WithString("location",
			tools.Description("The location to get weather for"),
			tools.Required(),
		),
	)
	plugin.AddTool(weatherTool, handleWeather)

	// Register routes with the PocketBase app
	// This is the key part to ensure proper integration
	basePath := plugin.GetBasePath()

	// CRITICAL: We need to ensure SSE and message endpoints are handled DIRECTLY
	// by their dedicated handlers, BEFORE any catch-all route has a chance

	// Get the raw SSE handler - this handler takes care of its own headers
	// and will properly implement the event stream protocol
	sseHandler := plugin.GetSSEHandler()
	app.RegisterRoute(basePath+"/sse", sseHandler)
	logger.Printf("Registered direct SSE endpoint: %s/sse (highest priority)", basePath)

	// Register specific endpoints
	app.RegisterRoute(basePath+"/message", plugin.GetHTTPHandler())
	app.RegisterRoute(basePath+"/tools", plugin.GetHTTPHandler())

	// Finally register the catch-all route for any other paths under the base path
	app.RegisterRoute(basePath+"/*", plugin.GetHTTPHandler())
	logger.Printf("Registered catch-all route: %s/*", basePath)

	// Start the server
	if err := app.Start(serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Printf("Server started on port %d", serverPort)
	log.Printf("MCP Service available at http://localhost:%d%s", serverPort, basePath)
	log.Printf("SSE endpoint: http://localhost:%d%s/sse", serverPort, basePath)
	log.Printf("Message endpoint: http://localhost:%d%s/message", serverPort, basePath)
	log.Printf("Tools endpoint: http://localhost:%d%s/tools", serverPort, basePath)
	log.Printf("Test client: http://localhost:%d/test-client.html", serverPort)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// Give 5 seconds for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// Echo tool handler
func handleEcho(ctx context.Context, request pocketbase.ToolCallRequest) (interface{}, error) {
	// Extract the message parameter
	message, ok := request.Parameters["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'message' parameter")
	}

	// Log that we received the request
	log.Printf("Echoing message: %s", message)

	// Return the echo response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}, nil
}

// Weather tool handler
func handleWeather(ctx context.Context, request pocketbase.ToolCallRequest) (interface{}, error) {
	// Extract the location parameter
	location, ok := request.Parameters["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'location' parameter")
	}

	// Log that we received the request
	log.Printf("Getting weather for: %s", location)

	// In a real app, we would call a weather API here
	// For this example, we'll just return a mock response
	weatherInfo := fmt.Sprintf("Weather for %s: 72°F, Partly Cloudy", location)

	// Return the weather response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": weatherInfo,
			},
		},
	}, nil
}
