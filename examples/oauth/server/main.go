package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

// This example demonstrates how to set up a Cortex server with OAuth 2.1 authentication

func main() {
	// Create a logger
	logger := log.New(os.Stderr, "[cortex] ", log.LstdFlags)

	// Set up the Cortex server
	mcpServer := server.NewMCPServer("OAuth2 Example Server", "1.0.0", logger)

	// Configure the server for HTTP
	mcpServer.SetAddress(":8080")

	// Register a simple tool to demonstrate
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Add the tool to the server
	mcpServer.AddTool(context.Background(), echoTool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		message := request.Parameters["message"].(string)
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		}, nil
	})

	// Create OAuth 2.1 configuration
	oauthConfig := &server.OAuthConfig{
		// The issuer URL for the auth server (e.g., Auth0, Keycloak, etc.)
		Issuer: "https://auth.example.com",

		// The expected audience value(s) in the token
		Audience: []string{"cortex-api"},

		// URL to the JSON Web Key Set for JWT validation
		JWKSUrl: "https://auth.example.com/.well-known/jwks.json",

		// Global scopes required for all requests
		RequiredScopes: []string{"cortex:api"},

		// How to extract tokens: "header" (default), "query", or "cookie"
		// Multiple sources can be specified with comma separation, e.g., "header,query"
		TokenLookupScheme: "header,query",

		// Header name for token extraction (default: "Authorization")
		TokenHeaderName: "Authorization",

		// Query parameter name for token extraction
		TokenQueryParam: "access_token",
	}

	// Create a key provider that fetches keys from the JWKS URL
	keyProvider := server.NewJWKSKeyProvider(oauthConfig.JWKSUrl)

	// Create a JWT token validator using the key provider
	tokenValidator := server.NewJWTTokenValidator(oauthConfig, keyProvider)

	// Create the OAuth middleware with the validator and configuration
	oauthMiddleware := server.NewOAuthMiddlewareWithConfig(tokenValidator, oauthConfig)

	// Create tool permissions manager
	toolPermissions := server.NewToolPermissions(oauthMiddleware)

	// Create an HTTP adapter for the MCP server
	adapter := server.NewHTTPAdapter(mcpServer, server.WithPath("/api/mcp"))

	// Create a router for the HTTP server
	mux := http.NewServeMux()

	// Add the MCP server to the router - wrap with OAuth middleware
	mux.Handle("/api/mcp/", oauthMiddleware.Middleware(adapter.Handler()))

	// Add public endpoints (not requiring auth)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Cortex OAuth2 Example Server"))
	})

	// Add protected endpoints with specific scope requirements
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a protected endpoint requiring the 'cortex:admin' scope"))
	})

	// Apply OAuth middleware with specific scope requirement
	mux.Handle("/admin", oauthMiddleware.RequireScope("cortex:admin", protectedHandler))

	// Create a protected endpoint for a specific tool with tool-specific permission
	toolHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Echo tool endpoint - requires execute permission for the echo tool"))
	})

	// Apply tool-specific permission middleware
	mux.Handle("/tools/echo", toolPermissions.RequireToolPermission(
		"echo",                       // Tool name
		server.ToolPermissionExecute, // Permission type (Execute, Read, Write)
		toolHandler,
	))

	// Start the HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		logger.Printf("Starting HTTP server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Graceful shutdown
	logger.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Printf("Server shutdown error: %v", err)
	}

	logger.Println("Server gracefully stopped")
}

// In a real application, you'd implement a proper token validator
// Below is a simple validator for demonstration purposes

// MockTokenValidator implements server.TokenValidator for testing
type MockTokenValidator struct{}

func NewMockTokenValidator() *MockTokenValidator {
	return &MockTokenValidator{}
}

func (v *MockTokenValidator) ValidateToken(ctx context.Context, token string) (*server.TokenClaims, error) {
	// In a real implementation, you would validate the token signature and claims
	// This example just validates a test token

	if token == "test-token" {
		// Return claims for a valid token
		return &server.TokenClaims{
			Subject:   "user123",
			Issuer:    "https://auth.example.com",
			Audience:  []string{"cortex-api"},
			ExpiresAt: time.Now().Add(time.Hour), // Token valid for 1 hour
			IssuedAt:  time.Now(),
			Scopes:    []string{"cortex:api", "cortex:tool:execute:echo"},
			Claims:    map[string]interface{}{},
		}, nil
	}

	// Return error for invalid token
	return nil, server.ErrInvalidToken
}

// To use the mock validator instead of JWT validator, replace the validator creation with:
// tokenValidator := NewMockTokenValidator()
// oauthMiddleware := server.NewOAuthMiddlewareWithConfig(tokenValidator, oauthConfig)
