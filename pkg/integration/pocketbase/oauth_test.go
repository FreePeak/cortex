package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FreePeak/cortex/pkg/server"
)

type mockTokenValidator struct {
	validateFunc func(ctx context.Context, token string) (*server.TokenClaims, error)
}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, token string) (*server.TokenClaims, error) {
	return m.validateFunc(ctx, token)
}

func TestOAuthMiddleware(t *testing.T) {
	// Create a plugin with default options
	plugin := NewCortexPlugin()

	// Create a mock token validator
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*server.TokenClaims, error) {
			if token == "valid-token" {
				return &server.TokenClaims{
					Subject:   "user123",
					Issuer:    "test-issuer",
					Audience:  []string{"test-audience"},
					ExpiresAt: time.Now().Add(time.Hour),
					IssuedAt:  time.Now(),
					Scopes:    []string{"cortex:tool:read", "cortex:tool:execute:echo"},
					Claims:    map[string]interface{}{},
				}, nil
			}
			return nil, server.ErrInvalidToken
		},
	}

	// Setup OAuth middleware
	plugin.WithOAuth(server.NewOAuthMiddleware(validator))

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that claims are present in context
		claims, ok := server.GetTokenClaimsFromContext(r.Context())
		assert.True(t, ok, "Claims should be in context")
		assert.Equal(t, "user123", claims.Subject)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create a test HTTP handler with middleware
	handler := plugin.GetOAuthHandler(nextHandler)

	// Test valid token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "success", recorder.Body.String())

	// Test invalid token
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	recorder = httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestWithOAuthConfig(t *testing.T) {
	// Create a plugin with default options
	plugin := NewCortexPlugin()

	// Create OAuth config
	config := &server.OAuthConfig{
		Issuer:            "https://auth.example.com",
		Audience:          []string{"cortex-api"},
		JWKSUrl:           "https://auth.example.com/.well-known/jwks.json",
		RequiredScopes:    []string{"cortex:api"},
		TokenLookupScheme: "header,query",
		TokenHeaderName:   "Authorization",
		TokenQueryParam:   "access_token",
	}

	// Apply OAuth config
	plugin = plugin.WithOAuthConfig(config)

	// Verify the config was applied
	assert.NotNil(t, plugin.oauthConfig)
	assert.Equal(t, "https://auth.example.com", plugin.oauthConfig.Issuer)
	assert.Equal(t, []string{"cortex-api"}, plugin.oauthConfig.Audience)
	assert.Equal(t, "https://auth.example.com/.well-known/jwks.json", plugin.oauthConfig.JWKSUrl)
	assert.Equal(t, []string{"cortex:api"}, plugin.oauthConfig.RequiredScopes)
	assert.Equal(t, "header,query", plugin.oauthConfig.TokenLookupScheme)
	assert.Equal(t, "Authorization", plugin.oauthConfig.TokenHeaderName)
	assert.Equal(t, "access_token", plugin.oauthConfig.TokenQueryParam)
}

func TestToolPermissionMiddleware(t *testing.T) {
	// Create a plugin with default options
	plugin := NewCortexPlugin()

	// Create a mock token validator
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*server.TokenClaims, error) {
			if token == "valid-token" {
				return &server.TokenClaims{
					Subject:   "user123",
					Issuer:    "test-issuer",
					Audience:  []string{"test-audience"},
					ExpiresAt: time.Now().Add(time.Hour),
					IssuedAt:  time.Now(),
					Scopes:    []string{"cortex:tool:read", "cortex:tool:execute:echo"},
					Claims:    map[string]interface{}{},
				}, nil
			}
			return nil, server.ErrInvalidToken
		},
	}

	// Setup OAuth middleware
	plugin.WithOAuth(server.NewOAuthMiddleware(validator))

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create a test HTTP handler with tool permission middleware
	handler := plugin.GetToolPermissionHandler("echo", server.ToolPermissionExecute, nextHandler)

	// Test valid token with correct permission
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "success", recorder.Body.String())

	// Test valid token with incorrect permission (different tool)
	handler = plugin.GetToolPermissionHandler("weather", server.ToolPermissionExecute, nextHandler)
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	recorder = httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}
