// Package server provides the MCP server implementation.
package server

import (
	"context"
	"net/http"
	"testing"
)

func TestTokenValidation(t *testing.T) {
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			if token == "valid-token" {
				return &TokenClaims{
					Subject: "user123",
					Scopes:  []string{"tools:read", "tools:execute"},
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	// Test valid token
	claims, err := validator.ValidateToken(context.Background(), "valid-token")
	if err != nil {
		t.Errorf("Expected no error for valid token, got %v", err)
	}
	if claims.Subject != "user123" {
		t.Errorf("Expected subject to be user123, got %s", claims.Subject)
	}
	if len(claims.Scopes) != 2 || claims.Scopes[0] != "tools:read" || claims.Scopes[1] != "tools:execute" {
		t.Errorf("Unexpected scopes: %v", claims.Scopes)
	}

	// Test invalid token
	_, err = validator.ValidateToken(context.Background(), "invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthMiddleware(t *testing.T) {
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			if token == "valid-token" {
				return &TokenClaims{
					Subject: "user123",
					Scopes:  []string{"tools:read", "tools:execute"},
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	middleware := NewOAuthMiddleware(validator)

	// Create a mock handler that will be wrapped by the middleware
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		// Check if TokenClaims were added to the context
		claims, ok := GetTokenClaimsFromContext(r.Context())
		if !ok {
			t.Error("TokenClaims not found in context")
			return
		}

		if claims.Subject != "user123" {
			t.Errorf("Expected subject to be user123, got %s", claims.Subject)
		}
	})

	handler := middleware.Middleware(next)

	// Test with valid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := &mockResponseWriter{}

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("Next handler was not called with valid token")
	}

	// Test with invalid token
	nextCalled = false
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = &mockResponseWriter{}

	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Error("Next handler was called with invalid token")
	}

	if rr.statusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.statusCode)
	}
}

// Mock implementations for testing

type mockTokenValidator struct {
	validateFunc func(ctx context.Context, token string) (*TokenClaims, error)
}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	return m.validateFunc(ctx, token)
}

type mockResponseWriter struct {
	statusCode int
	headers    http.Header
	body       []byte
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	m.body = b
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}
