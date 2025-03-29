// This file provides examples for integrating the logging package
// with the MCP server application.
package logging

import (
	"context"
	"net/http"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Defined context keys
const (
	loggerKey contextKey = "logger"
)

// Middleware creates an HTTP middleware that adds a logger to the request context.
func Middleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a child logger with request information
			requestLogger := logger.With(Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})

			// Add logger to context
			ctx := context.WithValue(r.Context(), loggerKey, requestLogger)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLogger retrieves the logger from the context.
// If no logger is found, returns a default logger.
func GetLogger(ctx context.Context) *Logger {
	logger, ok := ctx.Value(loggerKey).(*Logger)
	if !ok || logger == nil {
		// Return default logger if not found in context
		return Default()
	}
	return logger
}

// LogJSONRPCRequest logs JSON-RPC request details
func LogJSONRPCRequest(ctx context.Context, request domain.JSONRPCRequest) {
	logger := GetLogger(ctx)

	logger.Info("JSON-RPC Request",
		Fields{
			"id":      request.ID,
			"method":  request.Method,
			"jsonrpc": request.JSONRPC,
		})
}

// LogJSONRPCResponse logs JSON-RPC response details
func LogJSONRPCResponse(ctx context.Context, response domain.JSONRPCResponse) {
	logger := GetLogger(ctx)

	fields := Fields{
		"id":      response.ID,
		"jsonrpc": response.JSONRPC,
	}

	if response.Error != nil {
		fields["error_code"] = response.Error.Code
		fields["error_message"] = response.Error.Message
		logger.Error("JSON-RPC Response Error", fields)
	} else {
		logger.Info("JSON-RPC Response Success", fields)
	}
}

// ServerStartupLogger logs server startup information
func ServerStartupLogger(logger *Logger, serverName, version, address string) {
	logger.Info("Server starting",
		Fields{
			"name":    serverName,
			"version": version,
			"address": address,
		})
}

// WithRequestID returns a new logger with the request ID field
func WithRequestID(logger *Logger, requestID string) *Logger {
	return logger.With(Fields{
		"request_id": requestID,
	})
}
