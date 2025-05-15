// Package server provides the MCP server implementation.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OAuth 2.1 related errors
var (
	// ErrInvalidToken indicates that the provided token is invalid or expired
	ErrInvalidToken = errors.New("invalid or expired token")

	// ErrInsufficientScope indicates that the token does not have the required scope
	ErrInsufficientScope = errors.New("insufficient scope")

	// ErrMissingToken indicates that no token was provided
	ErrMissingToken = errors.New("missing token")

	// ErrInvalidRequest indicates an invalid OAuth request
	ErrInvalidRequest = errors.New("invalid request")
)

// contextKey is a custom type to use as keys in context.WithValue to avoid collisions
type contextKey string

// Define keys used in context
const (
	tokenClaimsContextKey contextKey = "tokenClaims"
)

// TokenClaims represents the validated claims from an OAuth 2.1 access token
type TokenClaims struct {
	// Subject is the user identifier
	Subject string

	// Issuer is the token issuer
	Issuer string

	// Audience contains the intended audience for this token
	Audience []string

	// ExpiresAt is the expiration time
	ExpiresAt time.Time

	// IssuedAt is when the token was issued
	IssuedAt time.Time

	// Scopes contains the OAuth scopes granted to this token
	Scopes []string

	// Additional claims can be added as needed
	Claims map[string]interface{}
}

// TokenValidator defines the interface for validating OAuth 2.1 tokens
type TokenValidator interface {
	// ValidateToken validates the provided token and returns the claims if valid
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

// ScopeChecker defines the interface for checking if a token has the required scopes
type ScopeChecker interface {
	// HasScope checks if the token has the required scope
	HasScope(claims *TokenClaims, requiredScope string) bool

	// HasAnyScope checks if the token has any of the required scopes
	HasAnyScope(claims *TokenClaims, requiredScopes ...string) bool

	// HasAllScopes checks if the token has all the required scopes
	HasAllScopes(claims *TokenClaims, requiredScopes ...string) bool
}

// OAuthMiddleware provides middleware for OAuth 2.1 authentication
type OAuthMiddleware struct {
	validator TokenValidator
	checker   ScopeChecker
}

// NewOAuthMiddleware creates a new OAuthMiddleware with the provided token validator
func NewOAuthMiddleware(validator TokenValidator) *OAuthMiddleware {
	return &OAuthMiddleware{
		validator: validator,
		checker:   &defaultScopeChecker{},
	}
}

// WithScopeChecker sets a custom scope checker for the middleware
func (m *OAuthMiddleware) WithScopeChecker(checker ScopeChecker) *OAuthMiddleware {
	m.checker = checker
	return m
}

// Middleware returns an http.Handler middleware that validates OAuth tokens
func (m *OAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.validator.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), tokenClaimsContextKey, claims)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope returns middleware that ensures the token has the required scope
func (m *OAuthMiddleware) RequireScope(requiredScope string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetTokenClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized: No token claims found", http.StatusUnauthorized)
			return
		}

		if !m.checker.HasScope(claims, requiredScope) {
			http.Error(w, fmt.Sprintf("Forbidden: Insufficient scope, requires %s", requiredScope), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAnyScope returns middleware that ensures the token has at least one of the required scopes
func (m *OAuthMiddleware) RequireAnyScope(requiredScopes []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetTokenClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized: No token claims found", http.StatusUnauthorized)
			return
		}

		if !m.checker.HasAnyScope(claims, requiredScopes...) {
			http.Error(w, fmt.Sprintf("Forbidden: Insufficient scope, requires one of %v", requiredScopes), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAllScopes returns middleware that ensures the token has all of the required scopes
func (m *OAuthMiddleware) RequireAllScopes(requiredScopes []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetTokenClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized: No token claims found", http.StatusUnauthorized)
			return
		}

		if !m.checker.HasAllScopes(claims, requiredScopes...) {
			http.Error(w, fmt.Sprintf("Forbidden: Insufficient scope, requires all of %v", requiredScopes), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetTokenClaimsFromContext extracts token claims from the context
func GetTokenClaimsFromContext(ctx context.Context) (*TokenClaims, bool) {
	claims, ok := ctx.Value(tokenClaimsContextKey).(*TokenClaims)
	return claims, ok
}

// defaultScopeChecker is the default implementation of ScopeChecker
type defaultScopeChecker struct{}

// HasScope checks if the token has the required scope
func (c *defaultScopeChecker) HasScope(claims *TokenClaims, requiredScope string) bool {
	for _, scope := range claims.Scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the token has any of the required scopes
func (c *defaultScopeChecker) HasAnyScope(claims *TokenClaims, requiredScopes ...string) bool {
	for _, requiredScope := range requiredScopes {
		if c.HasScope(claims, requiredScope) {
			return true
		}
	}
	return false
}

// HasAllScopes checks if the token has all of the required scopes
func (c *defaultScopeChecker) HasAllScopes(claims *TokenClaims, requiredScopes ...string) bool {
	for _, requiredScope := range requiredScopes {
		if !c.HasScope(claims, requiredScope) {
			return false
		}
	}
	return true
}

// OAuthConfig represents the configuration for OAuth 2.1 authorization
type OAuthConfig struct {
	// Issuer is the expected token issuer URL (iss claim)
	Issuer string

	// Audience is the expected audience (aud claim)
	Audience []string

	// JWKSUrl is the URL to the JSON Web Key Set for JWT validation
	JWKSUrl string

	// RequiredScopes are the scopes required for all requests
	RequiredScopes []string

	// TokenLookupScheme specifies how to extract tokens (e.g., "header", "query", "cookie")
	TokenLookupScheme string

	// TokenHeaderName is the name of the header for tokens (default: "Authorization")
	TokenHeaderName string

	// TokenQueryParam is the name of the query parameter for tokens
	TokenQueryParam string
}

// DefaultOAuthConfig returns a default configuration for OAuth 2.1
func DefaultOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		TokenLookupScheme: "header",
		TokenHeaderName:   "Authorization",
	}
}
