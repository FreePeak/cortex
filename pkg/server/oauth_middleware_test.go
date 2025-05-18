package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOAuthMiddlewareValidation(t *testing.T) {
	// Create a validator that accepts tokens in format "valid-token-{scope1}-{scope2}"
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			if len(token) > 12 && token[:12] == "valid-token-" {
				// Extract scopes from token for testing
				scopes := []string{}
				if len(token) > 12 {
					scope := token[12:]
					if scope != "" {
						scopes = append(scopes, scope)
					}
				}

				return &TokenClaims{
					Subject:   "test-user",
					Issuer:    "test-issuer",
					Audience:  []string{"test-audience"},
					ExpiresAt: time.Now().Add(time.Hour),
					IssuedAt:  time.Now(),
					Scopes:    scopes,
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	// Create middleware
	middleware := NewOAuthMiddleware(validator)

	// Test cases
	tests := []struct {
		name           string
		token          string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "Valid token",
			token:          "valid-token-read",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "Missing authorization header",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "Invalid token format",
			token:          "not-a-bearer-token",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test handler that will be called after the middleware
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Create the handler with our middleware
			handler := middleware.Middleware(next)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.token != "" {
				if tc.token == "not-a-bearer-token" {
					req.Header.Set("Authorization", tc.token)
				} else {
					req.Header.Set("Authorization", "Bearer "+tc.token)
				}
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rr, req)

			// Check if next handler was called
			if nextCalled != tc.shouldCallNext {
				t.Errorf("Expected next handler to be called: %v, but got: %v", tc.shouldCallNext, nextCalled)
			}

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}

			// If this was a valid token, verify claims are in context
			if tc.shouldCallNext {
				// We can't check context from the httptest recorder directly, so we'll test separately
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer "+tc.token)

				// Create a new handler that checks for context values
				contextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					claims, ok := GetTokenClaimsFromContext(r.Context())
					if !ok {
						t.Error("Claims not added to context")
						return
					}

					if claims.Subject != "test-user" {
						t.Errorf("Expected subject 'test-user', got %s", claims.Subject)
					}

					if len(claims.Scopes) == 0 || claims.Scopes[0] != "read" {
						t.Errorf("Expected scope 'read', got %v", claims.Scopes)
					}
				})

				// Create handler with middleware
				contextTestHandler := middleware.Middleware(contextHandler)

				// Call the handler
				contextTestHandler.ServeHTTP(httptest.NewRecorder(), req)
			}
		})
	}
}

func TestRequireScopeMiddleware(t *testing.T) {
	// Create validator
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			// Token format: valid-token-{scope}
			var scopes []string
			if token == "valid-token-read" {
				scopes = []string{"read"}
			} else if token == "valid-token-write" {
				scopes = []string{"write"}
			} else if token == "valid-token-admin" {
				scopes = []string{"admin"}
			} else if token == "valid-token-multiple" {
				scopes = []string{"read", "write"}
			}

			if len(scopes) > 0 {
				return &TokenClaims{
					Subject: "test-user",
					Scopes:  scopes,
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	// Create middleware
	middleware := NewOAuthMiddleware(validator)

	// Test cases
	tests := []struct {
		name           string
		token          string
		requiredScope  string
		expectedStatus int
	}{
		{
			name:           "Has required scope",
			token:          "valid-token-read",
			requiredScope:  "read",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing required scope",
			token:          "valid-token-read",
			requiredScope:  "write",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Multiple scopes - has required",
			token:          "valid-token-multiple",
			requiredScope:  "read",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test success handler
			successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create the two middleware layers:
			// 1. Scope requirement middleware
			// 2. Token validation middleware - this needs to run first to add claims to context
			scopeHandler := middleware.RequireScope(tc.requiredScope, successHandler)
			tokenHandler := middleware.Middleware(scopeHandler)

			// Create test request with token
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler chain
			tokenHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}

func TestMultipleScopeRequirements(t *testing.T) {
	// Create validator
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			if token == "valid-token" {
				return &TokenClaims{
					Subject: "test-user",
					Scopes:  []string{"read", "write", "admin"},
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	// Create middleware
	middleware := NewOAuthMiddleware(validator)

	// Create a test success handler
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test RequireAnyScope
	t.Run("RequireAnyScope", func(t *testing.T) {
		// Create scope middleware that requires any of the scopes
		scopeHandler := middleware.RequireAnyScope([]string{"delete", "write"}, successHandler)

		// Token handler adds the token to the context and should run first
		tokenHandler := middleware.Middleware(scopeHandler)

		// Create test request with valid token
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler chain
		tokenHandler.ServeHTTP(rr, req)

		// Should succeed because token has "write" scope
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		// Try with scopes the token doesn't have
		scopeHandler = middleware.RequireAnyScope([]string{"delete", "super-admin"}, successHandler)
		tokenHandler = middleware.Middleware(scopeHandler)
		rr = httptest.NewRecorder()
		tokenHandler.ServeHTTP(rr, req)

		// Should fail because token has none of the required scopes
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, rr.Code)
		}
	})

	// Test RequireAllScopes
	t.Run("RequireAllScopes", func(t *testing.T) {
		// Create scope middleware that requires all of the scopes
		scopeHandler := middleware.RequireAllScopes([]string{"read", "write"}, successHandler)

		// Token handler adds the token to the context and should run first
		tokenHandler := middleware.Middleware(scopeHandler)

		// Create test request with valid token
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler chain
		tokenHandler.ServeHTTP(rr, req)

		// Should succeed because token has both "read" and "write" scopes
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		// Try with a scope the token doesn't have
		scopeHandler = middleware.RequireAllScopes([]string{"read", "write", "delete"}, successHandler)
		tokenHandler = middleware.Middleware(scopeHandler)
		rr = httptest.NewRecorder()
		tokenHandler.ServeHTTP(rr, req)

		// Should fail because token doesn't have all required scopes
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, rr.Code)
		}
	})
}

// Additional test for token extraction from different sources
func TestTokenExtraction(t *testing.T) {
	// Create validator
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			if token == "valid-token" {
				return &TokenClaims{
					Subject: "test-user",
					Scopes:  []string{"read"},
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	// Create middleware with custom config
	config := &OAuthConfig{
		TokenLookupScheme: "header,query,cookie",
		TokenHeaderName:   "X-API-Key",
		TokenQueryParam:   "access_token",
	}

	t.Run("Create middleware with config", func(t *testing.T) {
		middleware := NewOAuthMiddlewareWithConfig(validator, config)

		if middleware == nil {
			t.Fatal("Middleware should not be nil")
		}

		// TODO: Implement the actual tests for token extraction
		// These will be implemented once we create the middleware implementation
	})
}
