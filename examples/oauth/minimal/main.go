package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
)

// This example demonstrates a minimal OAuth 2.1 setup with Cortex
// It only shows the OAuth-specific parts, not the full server setup

func main() {
	// 1. Create OAuth 2.1 configuration
	oauthConfig := &server.OAuthConfig{
		Issuer:            "https://auth.example.com",
		Audience:          []string{"cortex-api"},
		JWKSUrl:           "https://auth.example.com/.well-known/jwks.json",
		RequiredScopes:    []string{"cortex:api"},
		TokenLookupScheme: "header,query",
		TokenHeaderName:   "Authorization",
		TokenQueryParam:   "access_token",
	}

	// 2. Set up token validation - choose one approach:

	// A. JWT Validation (recommended for production)
	keyProvider := server.NewJWKSKeyProvider(oauthConfig.JWKSUrl)
	jwtValidator := server.NewJWTTokenValidator(oauthConfig, keyProvider)

	// B. Token Introspection (for opaque tokens)
	// introspector := server.NewStandardIntrospector(
	//     "https://auth.example.com/oauth/introspect",
	//     "client_id",
	//     "client_secret",
	// )
	// introspectionValidator := server.NewIntrospectionTokenValidator(oauthConfig, introspector)

	// C. For testing: use a simple mock validator
	// mockValidator := &SimpleMockTokenValidator{}

	// 3. Create OAuth middleware with the chosen validator
	oauthMiddleware := server.NewOAuthMiddlewareWithConfig(jwtValidator, oauthConfig)

	// 4. Create handlers using the middleware

	// Basic protected endpoint - requires a valid token
	http.Handle("/protected", oauthMiddleware.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token claims from the request context
			claims, ok := server.GetTokenClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Use claims in your response
			response := map[string]interface{}{
				"message": "You have access to the protected resource",
				"userId":  claims.Subject,
				"scopes":  claims.Scopes,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}),
	))

	// Scope-specific endpoint - requires the "cortex:admin" scope
	http.Handle("/admin", oauthMiddleware.RequireScope("cortex:admin",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Admin access granted"))
		}),
	))

	// Endpoint requiring multiple scopes
	http.Handle("/high-privilege", oauthMiddleware.RequireAllScopes(
		[]string{"cortex:admin", "cortex:high-privilege"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("High privilege access granted"))
		}),
	))

	// 5. Start the server
	log.Println("Starting OAuth 2.1 example server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// SimpleMockTokenValidator is a simple token validator for testing
type SimpleMockTokenValidator struct{}

func (m *SimpleMockTokenValidator) ValidateToken(ctx context.Context, token string) (*server.TokenClaims, error) {
	// For testing: accept a specific token
	if token == "test-token" {
		return &server.TokenClaims{
			Subject:   "test-user",
			Issuer:    "https://auth.example.com",
			Audience:  []string{"cortex-api"},
			ExpiresAt: time.Now().Add(time.Hour),
			IssuedAt:  time.Now(),
			Scopes:    []string{"cortex:api", "cortex:admin"},
			Claims:    map[string]interface{}{},
		}, nil
	}
	return nil, server.ErrInvalidToken
}
