// This file provides examples for integrating the logging package
// with the MCP server application.
package logging

import (
	"context"
	"net/http"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain"
)

// LoggingMiddleware creates an HTTP middleware that logs requests
func LoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create context with logger
			ctx := context.WithValue(r.Context(), "logger", logger)

			// Call the next handler with the enhanced context
			next.ServeHTTP(w, r.WithContext(ctx))

			// Log request details
			logger.Info("HTTP Request",
				Fields{
					"method":      r.Method,
					"path":        r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"user_agent":  r.UserAgent(),
					"duration_ms": time.Since(start).Milliseconds(),
				})
		})
	}
}

// GetLoggerFromContext extracts logger from context
func GetLoggerFromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value("logger").(*Logger); ok {
		return logger
	}
	// Return default logger if not found in context
	return Default()
}

// LogJSONRPCRequest logs JSON-RPC request details
func LogJSONRPCRequest(ctx context.Context, request domain.JSONRPCRequest) {
	logger := GetLoggerFromContext(ctx)

	logger.Info("JSON-RPC Request",
		Fields{
			"id":      request.ID,
			"method":  request.Method,
			"jsonrpc": request.JSONRPC,
		})
}

// LogJSONRPCResponse logs JSON-RPC response details
func LogJSONRPCResponse(ctx context.Context, response domain.JSONRPCResponse) {
	logger := GetLoggerFromContext(ctx)

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
