package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestToolPermissionMiddleware(t *testing.T) {
	// Create a test server with middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		toolName       string
		permission     ToolPermission
		expectedStatus int
		scopes         []string
	}{
		{
			name:           "Allow Execute Permission With Proper Scope",
			toolName:       "calculator",
			permission:     ToolPermissionExecute,
			expectedStatus: http.StatusOK,
			scopes:         []string{"cortex:tool:execute:calculator"},
		},
		{
			name:           "Deny Execute Permission Without Proper Scope",
			toolName:       "weather",
			permission:     ToolPermissionExecute,
			expectedStatus: http.StatusForbidden,
			scopes:         []string{"cortex:tool:execute:calculator"},
		},
		{
			name:           "Allow Read Permission With Global Read Scope",
			toolName:       "any-tool",
			permission:     ToolPermissionRead,
			expectedStatus: http.StatusOK,
			scopes:         []string{"cortex:tool:read"},
		},
		{
			name:           "Allow Write Permission With Proper Scope",
			toolName:       "calculator",
			permission:     ToolPermissionWrite,
			expectedStatus: http.StatusOK,
			scopes:         []string{"cortex:tool:write:calculator"},
		},
		{
			name:           "Deny Write Permission Without Proper Scope",
			toolName:       "database",
			permission:     ToolPermissionWrite,
			expectedStatus: http.StatusForbidden,
			scopes:         []string{"cortex:tool:write:calculator"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create validator for this test
			testValidator := &mockTokenValidator{
				validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
					return &TokenClaims{
						Subject:   "user123",
						Issuer:    "test-issuer",
						Audience:  []string{"test-audience"},
						ExpiresAt: time.Now().Add(time.Hour),
						IssuedAt:  time.Now(),
						Scopes:    tc.scopes,
						Claims:    map[string]interface{}{},
					}, nil
				},
			}

			testOAuthMiddleware := NewOAuthMiddleware(testValidator)
			testToolPermissions := NewToolPermissions(testOAuthMiddleware)

			// Create handler with middleware
			handler := testOAuthMiddleware.Middleware(
				testToolPermissions.RequireToolPermission(tc.toolName, tc.permission, nextHandler),
			)

			// Create a test request
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Execute the request
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			// Check response
			if recorder.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestToolPermissionScopeFormat(t *testing.T) {
	// Test the formatting of tool permission scopes
	permissions := NewToolPermissions(nil)

	tests := []struct {
		name          string
		toolName      string
		permission    ToolPermission
		expectedScope string
	}{
		{
			name:          "Execute Permission",
			toolName:      "calculator",
			permission:    ToolPermissionExecute,
			expectedScope: "cortex:tool:execute:calculator",
		},
		{
			name:          "Read Permission",
			toolName:      "database",
			permission:    ToolPermissionRead,
			expectedScope: "cortex:tool:read:database",
		},
		{
			name:          "Write Permission",
			toolName:      "file-system",
			permission:    ToolPermissionWrite,
			expectedScope: "cortex:tool:write:file-system",
		},
		{
			name:          "Global Read Permission",
			toolName:      "",
			permission:    ToolPermissionRead,
			expectedScope: "cortex:tool:read",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope := permissions.FormatToolScope(tc.toolName, tc.permission)
			if scope != tc.expectedScope {
				t.Errorf("Expected scope %s, got %s", tc.expectedScope, scope)
			}
		})
	}
}
