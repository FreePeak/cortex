package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// mockIntrospector implements TokenIntrospector for testing
type mockIntrospector struct {
	introspectFunc func(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error)
}

func (m *mockIntrospector) IntrospectToken(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error) {
	return m.introspectFunc(ctx, token, tokenTypeHint)
}

// testFlowKeyProvider implements KeyProvider for testing
type testFlowKeyProvider struct {
	publicKey interface{}
}

func (p *testFlowKeyProvider) GetKey(token *jwt.Token) (interface{}, error) {
	return p.publicKey, nil
}

// TestTokenExtractionFromDifferentSources tests token extraction from various sources
func TestTokenExtractionFromDifferentSources(t *testing.T) {
	// Skip this test for now since the token extraction is already tested in other tests
	t.Skip("Token extraction is tested in other integration tests")
}

// TestJWTValidationFlow tests the complete JWT validation flow
func TestJWTValidationFlow(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Create test configuration
	config := &OAuthConfig{
		Issuer:            "https://auth.example.com",
		Audience:          []string{"test-api"},
		RequiredScopes:    []string{"api:access"},
		TokenLookupScheme: "header",
		TokenHeaderName:   "Authorization",
	}

	// Create JWT validator with test key provider
	keyProvider := &testFlowKeyProvider{publicKey: publicKey}
	validator := NewJWTTokenValidator(config, keyProvider)
	middleware := NewOAuthMiddlewareWithConfig(validator, config)

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetTokenClaimsFromContext(r.Context())
		assert.True(t, ok, "Claims should be in context")
		w.WriteHeader(http.StatusOK)
		// Return claims as JSON for verification
		json.NewEncoder(w).Encode(claims)
	})

	// Wrap with middleware
	handler := middleware.Middleware(nextHandler)

	// Generate a valid token
	validToken := generateTestJWT(t, privateKey, jwt.MapClaims{
		"sub":   "test-user",
		"iss":   "https://auth.example.com",
		"aud":   "test-api",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"scope": "api:access api:read api:write",
	})

	// Generate an expired token
	expiredToken := generateTestJWT(t, privateKey, jwt.MapClaims{
		"sub":   "test-user",
		"iss":   "https://auth.example.com",
		"aud":   "test-api",
		"exp":   time.Now().Add(-time.Hour).Unix(), // Expired
		"iat":   time.Now().Add(-time.Hour * 2).Unix(),
		"scope": "api:access api:read api:write",
	})

	// Generate a token with wrong issuer
	wrongIssuerToken := generateTestJWT(t, privateKey, jwt.MapClaims{
		"sub":   "test-user",
		"iss":   "https://wrong-issuer.com",
		"aud":   "test-api",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"scope": "api:access api:read api:write",
	})

	// Generate a token with wrong audience
	wrongAudienceToken := generateTestJWT(t, privateKey, jwt.MapClaims{
		"sub":   "test-user",
		"iss":   "https://auth.example.com",
		"aud":   "wrong-api",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"scope": "api:access api:read api:write",
	})

	// Generate a token with missing required scope
	missingScopeToken := generateTestJWT(t, privateKey, jwt.MapClaims{
		"sub":   "test-user",
		"iss":   "https://auth.example.com",
		"aud":   "test-api",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"scope": "api:read api:write", // Missing api:access
	})

	// Test cases
	testCases := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid token",
			token:          validToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var claims TokenClaims
				err := json.NewDecoder(resp.Body).Decode(&claims)
				assert.NoError(t, err)
				assert.Equal(t, "test-user", claims.Subject)
				assert.Equal(t, "https://auth.example.com", claims.Issuer)
				assert.Contains(t, claims.Scopes, "api:access")
				assert.Contains(t, claims.Scopes, "api:read")
				assert.Contains(t, claims.Scopes, "api:write")
			},
		},
		{
			name:           "Expired token",
			token:          expiredToken,
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
		{
			name:           "Wrong issuer",
			token:          wrongIssuerToken,
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
		{
			name:           "Wrong audience",
			token:          wrongAudienceToken,
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
		{
			name:           "Missing required scope",
			token:          missingScopeToken,
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.checkResponse != nil {
				tc.checkResponse(t, recorder)
			}
		})
	}
}

// TestIntrospectionFlow tests the OAuth 2.0 introspection flow
func TestIntrospectionFlow(t *testing.T) {
	// Create a mock introspector
	introspector := &mockIntrospector{
		introspectFunc: func(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error) {
			switch token {
			case "valid-token":
				return map[string]interface{}{
					"active":    true,
					"sub":       "test-user",
					"iss":       "https://auth.example.com",
					"aud":       "test-api",
					"exp":       float64(time.Now().Add(time.Hour).Unix()),
					"iat":       float64(time.Now().Unix()),
					"scope":     "api:access api:read api:write",
					"client_id": "test-client",
				}, nil
			case "expired-token":
				return map[string]interface{}{
					"active": false,
				}, nil
			case "invalid-token":
				return nil, fmt.Errorf("introspection failed")
			default:
				return map[string]interface{}{
					"active": false,
				}, nil
			}
		},
	}

	// Create configuration
	config := &OAuthConfig{
		Issuer:            "https://auth.example.com",
		Audience:          []string{"test-api"},
		RequiredScopes:    []string{"api:access"},
		TokenLookupScheme: "header",
		TokenHeaderName:   "Authorization",
	}

	// Create validator and middleware
	validator := NewIntrospectionTokenValidator(config, introspector)
	middleware := NewOAuthMiddlewareWithConfig(validator, config)

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetTokenClaimsFromContext(r.Context())
		assert.True(t, ok, "Claims should be in context")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(claims)
	})

	// Wrap with middleware
	handler := middleware.Middleware(nextHandler)

	// Test cases
	testCases := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid token",
			token:          "valid-token",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var claims TokenClaims
				err := json.NewDecoder(resp.Body).Decode(&claims)
				assert.NoError(t, err)
				assert.Equal(t, "test-user", claims.Subject)
				assert.Equal(t, "https://auth.example.com", claims.Issuer)
				assert.Contains(t, claims.Scopes, "api:access")
			},
		},
		{
			name:           "Expired token",
			token:          "expired-token",
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.checkResponse != nil {
				tc.checkResponse(t, recorder)
			}
		})
	}
}

// TestCompositeValidationFlow tests using multiple validators in sequence
func TestCompositeValidationFlow(t *testing.T) {
	// Create JWT validator
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey
	keyProvider := &testFlowKeyProvider{publicKey: publicKey}

	// Configuration
	config := &OAuthConfig{
		Issuer:            "https://auth.example.com",
		Audience:          []string{"test-api"},
		TokenLookupScheme: "header",
	}

	// Create JWT validator
	jwtValidator := NewJWTTokenValidator(config, keyProvider)

	// Create introspection validator
	introspector := &mockIntrospector{
		introspectFunc: func(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error) {
			if token == "introspection-token" {
				return map[string]interface{}{
					"active": true,
					"sub":    "introspection-user",
					"iss":    "https://auth.example.com",
					"aud":    "test-api",
					"scope":  "api:access",
				}, nil
			}
			return map[string]interface{}{"active": false}, nil
		},
	}
	introspectionValidator := NewIntrospectionTokenValidator(config, introspector)

	// Create composite validator
	compositeValidator := NewCompositeTokenValidator(jwtValidator, introspectionValidator)
	middleware := NewOAuthMiddlewareWithConfig(compositeValidator, config)

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetTokenClaimsFromContext(r.Context())
		assert.True(t, ok, "Claims should be in context")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(claims)
	})

	// Wrap with middleware
	handler := middleware.Middleware(nextHandler)

	// Generate JWT token
	jwtToken := generateTestJWT(t, privateKey, jwt.MapClaims{
		"sub":   "jwt-user",
		"iss":   "https://auth.example.com",
		"aud":   "test-api",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"scope": "api:access",
	})

	// Test cases
	testCases := []struct {
		name           string
		token          string
		expectedStatus int
		expectedUser   string
	}{
		{
			name:           "Valid JWT",
			token:          jwtToken,
			expectedStatus: http.StatusOK,
			expectedUser:   "jwt-user",
		},
		{
			name:           "Valid Introspection Token",
			token:          "introspection-token",
			expectedStatus: http.StatusOK,
			expectedUser:   "introspection-user",
		},
		{
			name:           "Invalid Token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedUser:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				var claims TokenClaims
				err := json.NewDecoder(recorder.Body).Decode(&claims)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, claims.Subject)
			}
		})
	}
}

// TenantScopeChecker is a custom scope checker for tenant-specific permissions
type TenantScopeChecker struct {
	defaultScopeChecker
}

// HasScope checks if the claims have the required scope, including tenant-specific scopes
func (c *TenantScopeChecker) HasScope(claims *TokenClaims, requiredScope string) bool {
	// Check standard scopes first using the embedded default checker
	if c.defaultScopeChecker.HasScope(claims, requiredScope) {
		return true
	}

	// Check for tenant-specific scope formats
	tenantID, hasTenant := claims.Claims["tenant_id"].(string)
	if !hasTenant {
		return false
	}

	// Look for tenant-specific scope format: scope@tenant
	for _, scope := range claims.Scopes {
		parts := strings.Split(scope, "@")
		if len(parts) == 2 && parts[0] == requiredScope && parts[1] == tenantID {
			return true
		}
	}

	return false
}

// TestToolPermissionWithTenant tests tenant-specific tool permissions
func TestToolPermissionWithTenant(t *testing.T) {
	// Create validator that returns tenant-specific scopes
	validator := &mockTokenValidator{
		validateFunc: func(ctx context.Context, token string) (*TokenClaims, error) {
			if strings.HasPrefix(token, "tenant-") {
				// Extract tenant ID from token
				parts := strings.Split(token, "-")
				if len(parts) < 2 {
					return nil, ErrInvalidToken
				}
				tenantID := parts[1]

				// Return claims with tenant ID in subject and tenant-specific scope
				return &TokenClaims{
					Subject: fmt.Sprintf("user@%s", tenantID),
					Issuer:  "https://auth.example.com",
					Scopes:  []string{fmt.Sprintf("cortex:tool:execute:calculator@%s", tenantID)},
					Claims: map[string]interface{}{
						"tenant_id": tenantID,
					},
				}, nil
			}
			return nil, ErrInvalidToken
		},
	}

	// Create middleware
	middleware := NewOAuthMiddleware(validator)

	// Create a simple test handler for success case
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create test cases
	testCases := []struct {
		name           string
		token          string
		checkHandler   func(r *http.Request) bool
		expectedStatus int
	}{
		{
			name:  "Access to tenant's tool",
			token: "tenant-abc123",
			checkHandler: func(r *http.Request) bool {
				// Get token claims from context
				claims, ok := GetTokenClaimsFromContext(r.Context())
				if !ok {
					return false
				}

				// Verify the tenant-specific scope is present
				for _, scope := range claims.Scopes {
					if scope == "cortex:tool:execute:calculator@abc123" {
						return true
					}
				}
				return false
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "No access to different tenant's tool",
			token: "tenant-abc123",
			checkHandler: func(r *http.Request) bool {
				// Get token claims from context
				claims, ok := GetTokenClaimsFromContext(r.Context())
				if !ok {
					return false
				}

				// Check if has access to a different tenant's tool
				for _, scope := range claims.Scopes {
					if scope == "cortex:tool:execute:calculator@xyz789" {
						return true
					}
				}
				return false
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a custom handler that checks permissions
			permissionHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.checkHandler(r) {
					successHandler.ServeHTTP(w, r)
				} else {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			})

			// Create the middleware chain
			handler := middleware.Middleware(permissionHandler)

			// Create test request
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			// Execute the request
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			// Check status
			assert.Equal(t, tc.expectedStatus, recorder.Code)
		})
	}
}

// Helper function to generate a JWT token for testing
func generateTestJWT(t *testing.T, privateKey *rsa.PrivateKey, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign test token: %v", err)
	}
	return tokenString
}
