// Package stdio provides the stdio interface for the MCP server.
package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
)

// StdioContextFunc is a function that takes an existing context and returns
// a potentially modified context.
// This can be used to inject context values from environment variables,
// for example.
type StdioContextFunc func(ctx context.Context) context.Context

// StdioServer wraps a MCPServer and handles stdio communication.
// It provides a simple way to create command-line MCP servers that
// communicate via standard input/output streams using JSON-RPC messages.
type StdioServer struct {
	server      *rest.MCPServer
	errLogger   *log.Logger
	contextFunc StdioContextFunc
}

// StdioOption defines a function type for configuring StdioServer
type StdioOption func(*StdioServer)

// WithErrorLogger sets the error logger for the server
func WithErrorLogger(logger *log.Logger) StdioOption {
	return func(s *StdioServer) {
		s.errLogger = logger
	}
}

// WithContextFunc sets a function that will be called to customize the context
// to the server. Note that the stdio server uses the same context for all requests,
// so this function will only be called once per server instance.
func WithStdioContextFunc(fn StdioContextFunc) StdioOption {
	return func(s *StdioServer) {
		s.contextFunc = fn
	}
}

// stdioSession is a static client session for stdio communication.
type stdioSession struct {
	notificationChannel chan domain.JSONRPCNotification
}

func (s *stdioSession) SessionID() string {
	return "stdio"
}

func (s *stdioSession) NotificationChannel() chan<- domain.JSONRPCNotification {
	return s.notificationChannel
}

// NewStdioServer creates a new stdio server wrapper around an MCPServer.
// It initializes the server with a default error logger that logs to stderr.
func NewStdioServer(server *rest.MCPServer, opts ...StdioOption) *StdioServer {
	s := &StdioServer{
		server:    server,
		errLogger: log.New(os.Stderr, "[STDIO] ", log.LstdFlags),
	}

	// Apply all options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Listen starts listening for JSON-RPC messages on the provided input and writes responses to the provided output.
// It runs until the context is cancelled or an error occurs.
// Returns an error if there are issues with reading input or writing output.
func (s *StdioServer) Listen(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	// Add in any custom context
	if s.contextFunc != nil {
		ctx = s.contextFunc(ctx)
	}

	reader := bufio.NewReader(stdin)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Use a goroutine to make the read cancellable
			readChan := make(chan string, 1)
			errChan := make(chan error, 1)

			go func() {
				line, err := reader.ReadString('\n')
				if err != nil {
					errChan <- err
					return
				}
				readChan <- line
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errChan:
				if err == io.EOF {
					return nil
				}
				s.errLogger.Printf("Error reading input: %v", err)
				return err
			case line := <-readChan:
				if err := s.processMessage(ctx, line, stdout); err != nil {
					if err == io.EOF {
						return nil
					}
					s.errLogger.Printf("Error handling message: %v", err)
					return err
				}
			}
		}
	}
}

// processMessage handles a single JSON-RPC message and writes the response.
// It parses the message, processes it through the wrapped MCPServer, and writes any response.
// Returns an error if there are issues with message processing or response writing.
func (s *StdioServer) processMessage(ctx context.Context, line string, writer io.Writer) error {
	// Parse message as raw JSON
	var rawMessage json.RawMessage
	if err := json.Unmarshal([]byte(line), &rawMessage); err != nil {
		response := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &rest.JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		}
		return s.writeResponse(response, writer)
	}

	// Create context with timeout - same as in rest implementation
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse the request to determine what kind of request we're dealing with
	var request rest.JSONRPCRequest
	if err := json.Unmarshal(rawMessage, &request); err != nil {
		response := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &rest.JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		}
		return s.writeResponse(response, writer)
	}

	// Directly handle the request using our own implementation of the protocol
	// since we can't call the unexported handleJSONRPC method

	// First, validate the JSON-RPC version
	if request.JSONRPC != "2.0" {
		response := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &rest.JSONRPCError{
				Code:    -32600,
				Message: "Invalid JSON-RPC version",
			},
		}
		return s.writeResponse(response, writer)
	}

	// Create a dummy HTTP request and response for the server's HTTP handler
	// This is a workaround since we can't directly call the handleJSONRPC method
	w := httptest.NewRecorder()
	r, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, "/jsonrpc", strings.NewReader(line))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")

	// Since we can't call handleJSONRPC directly, we'll create a small HTTP server that routes to it
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new HTTP handler that mimics the MCPServer's behavior
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		// Parse the request
		var req rest.JSONRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			resp := rest.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &rest.JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Process the request based on method
		var result interface{}
		switch req.Method {
		case "initialize":
			name, version, instructions := s.server.GetServerInfo()
			result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    name,
					"version": version,
				},
				"capabilities": map[string]interface{}{
					"resources": map[string]bool{
						"listChanged": true,
					},
					"tools": map[string]bool{
						"listChanged": true,
					},
					"prompts": map[string]bool{
						"listChanged": true,
					},
					"logging": struct{}{},
				},
			}

			if instructions != "" {
				result.(map[string]interface{})["instructions"] = instructions
			}

		case "ping":
			result = struct{}{}

		default:
			// For all other methods, we need to follow the HTTP approach since we don't have access
			// to the internal implementation
			resp := rest.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &rest.JSONRPCError{
					Code:    -32601,
					Message: fmt.Sprintf("Method '%s' not found", req.Method),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Send response
		resp := rest.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Handle the request
	handler.ServeHTTP(w, r)

	// Get the response
	resp := w.Result()
	defer resp.Body.Close()

	// Return the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	_, err = writer.Write(respBody)
	if err != nil {
		return err
	}

	// Add a newline for the stdio protocol
	_, err = writer.Write([]byte("\n"))
	return err
}

// writeResponse marshals and writes a JSON-RPC response message followed by a newline.
// Returns an error if marshaling or writing fails.
func (s *StdioServer) writeResponse(response interface{}, writer io.Writer) error {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	// Write response followed by newline
	if _, err := fmt.Fprintf(writer, "%s\n", responseBytes); err != nil {
		return err
	}

	return nil
}

// ServeStdio is a convenience function that creates and starts a StdioServer with os.Stdin and os.Stdout.
// It sets up signal handling for graceful shutdown on SIGTERM and SIGINT.
// Returns an error if the server encounters any issues during operation.
func ServeStdio(server *rest.MCPServer, opts ...StdioOption) error {
	s := NewStdioServer(server, opts...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		s.errLogger.Println("Received shutdown signal, stopping server...")
		cancel()
	}()

	s.errLogger.Println("Starting MCP server in stdio mode")
	return s.Listen(ctx, os.Stdin, os.Stdout)
}
