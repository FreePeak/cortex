// Package server provides the MCP server implementation.
package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTTokenValidator(t *testing.T) {
	// Create test private key for signing tokens
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Create validator with test configuration
	config := &OAuthConfig{
		Issuer:   "https://test-issuer.example.com",
		Audience: []string{"test-client"},
	}
	validator := NewJWTTokenValidator(config, &testKeyProvider{publicKey: &privateKey.PublicKey})

	// Test cases
	testCases := []struct {
		name           string
		tokenFunc      func() string
		expectedError  error
		expectedClaims *TokenClaims
	}{
		{
			name: "Valid token",
			tokenFunc: func() string {
				// Create valid claims
				claims := jwt.MapClaims{
					"sub":   "user123",
					"iss":   "https://test-issuer.example.com",
					"aud":   "test-client",
					"exp":   time.Now().Add(time.Hour).Unix(),
					"iat":   time.Now().Unix(),
					"scope": "tools:read tools:execute",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			expectedError: nil,
			expectedClaims: &TokenClaims{
				Subject:  "user123",
				Issuer:   "https://test-issuer.example.com",
				Audience: []string{"test-client"},
				Scopes:   []string{"tools:read", "tools:execute"},
			},
		},
		{
			name: "Expired token",
			tokenFunc: func() string {
				// Create expired claims
				claims := jwt.MapClaims{
					"sub":   "user123",
					"iss":   "https://test-issuer.example.com",
					"aud":   "test-client",
					"exp":   time.Now().Add(-time.Hour).Unix(), // Expired
					"iat":   time.Now().Add(-time.Hour * 2).Unix(),
					"scope": "tools:read tools:execute",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			expectedError:  ErrInvalidToken,
			expectedClaims: nil,
		},
		{
			name: "Invalid issuer",
			tokenFunc: func() string {
				// Create claims with wrong issuer
				claims := jwt.MapClaims{
					"sub":   "user123",
					"iss":   "https://wrong-issuer.example.com",
					"aud":   "test-client",
					"exp":   time.Now().Add(time.Hour).Unix(),
					"iat":   time.Now().Unix(),
					"scope": "tools:read tools:execute",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			expectedError:  ErrInvalidToken,
			expectedClaims: nil,
		},
		{
			name: "Invalid audience",
			tokenFunc: func() string {
				// Create claims with wrong audience
				claims := jwt.MapClaims{
					"sub":   "user123",
					"iss":   "https://test-issuer.example.com",
					"aud":   "wrong-client",
					"exp":   time.Now().Add(time.Hour).Unix(),
					"iat":   time.Now().Unix(),
					"scope": "tools:read tools:execute",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				return tokenString
			},
			expectedError:  ErrInvalidToken,
			expectedClaims: nil,
		},
		{
			name: "Invalid signature",
			tokenFunc: func() string {
				// Create valid token
				claims := jwt.MapClaims{
					"sub":   "user123",
					"iss":   "https://test-issuer.example.com",
					"aud":   "test-client",
					"exp":   time.Now().Add(time.Hour).Unix(),
					"iat":   time.Now().Unix(),
					"scope": "tools:read tools:execute",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString(privateKey)
				// Tamper with the token
				return tokenString + "invalid"
			},
			expectedError:  ErrInvalidToken,
			expectedClaims: nil,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.tokenFunc()
			claims, err := validator.ValidateToken(context.Background(), token)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tc.expectedClaims.Subject, claims.Subject)
				assert.Equal(t, tc.expectedClaims.Issuer, claims.Issuer)
				assert.ElementsMatch(t, tc.expectedClaims.Audience, claims.Audience)
				assert.ElementsMatch(t, tc.expectedClaims.Scopes, claims.Scopes)
			}
		})
	}
}

func TestIntrospectionTokenValidator(t *testing.T) {
	// Create mock introspection server
	mockServer := &mockIntrospectionServer{
		responses: map[string]introspectionResponse{
			"valid-token": {
				Active:   true,
				Subject:  "user123",
				Issuer:   "https://test-issuer.example.com",
				Audience: []string{"test-client"},
				Scope:    "tools:read tools:execute",
			},
			"expired-token": {
				Active: false,
			},
			"insufficient-scope-token": {
				Active:   true,
				Subject:  "user123",
				Issuer:   "https://test-issuer.example.com",
				Audience: []string{"test-client"},
				Scope:    "tools:read",
			},
		},
	}

	// Create validator with test configuration
	config := &OAuthConfig{
		Issuer:   "https://test-issuer.example.com",
		Audience: []string{"test-client"},
	}
	validator := NewIntrospectionTokenValidator(config, mockServer)

	// Test cases
	testCases := []struct {
		name           string
		token          string
		expectedError  error
		expectedClaims *TokenClaims
	}{
		{
			name:          "Valid token",
			token:         "valid-token",
			expectedError: nil,
			expectedClaims: &TokenClaims{
				Subject:  "user123",
				Issuer:   "https://test-issuer.example.com",
				Audience: []string{"test-client"},
				Scopes:   []string{"tools:read", "tools:execute"},
			},
		},
		{
			name:           "Expired token",
			token:          "expired-token",
			expectedError:  ErrInvalidToken,
			expectedClaims: nil,
		},
		{
			name:          "Insufficient scope",
			token:         "insufficient-scope-token",
			expectedError: nil,
			expectedClaims: &TokenClaims{
				Subject:  "user123",
				Issuer:   "https://test-issuer.example.com",
				Audience: []string{"test-client"},
				Scopes:   []string{"tools:read"},
			},
		},
		{
			name:           "Unknown token",
			token:          "unknown-token",
			expectedError:  ErrInvalidToken,
			expectedClaims: nil,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := validator.ValidateToken(context.Background(), tc.token)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tc.expectedClaims.Subject, claims.Subject)
				assert.Equal(t, tc.expectedClaims.Issuer, claims.Issuer)
				assert.ElementsMatch(t, tc.expectedClaims.Audience, claims.Audience)
				assert.ElementsMatch(t, tc.expectedClaims.Scopes, claims.Scopes)
			}
		})
	}
}

// Test helpers

type testKeyProvider struct {
	publicKey *rsa.PublicKey
}

func (p *testKeyProvider) GetKey(token *jwt.Token) (interface{}, error) {
	return p.publicKey, nil
}

type introspectionResponse struct {
	Active   bool     `json:"active"`
	Subject  string   `json:"sub"`
	Issuer   string   `json:"iss"`
	Audience []string `json:"aud"`
	Scope    string   `json:"scope"`
}

type mockIntrospectionServer struct {
	responses map[string]introspectionResponse
}

func (s *mockIntrospectionServer) IntrospectToken(ctx context.Context, token string, tokenTypeHint string) (map[string]interface{}, error) {
	response, exists := s.responses[token]
	if !exists {
		return map[string]interface{}{"active": false}, nil
	}

	result := map[string]interface{}{
		"active": response.Active,
	}

	if response.Active {
		result["sub"] = response.Subject
		result["iss"] = response.Issuer

		// Fix: Store audience as a string to match our validator's expectation
		if len(response.Audience) > 0 {
			result["aud"] = response.Audience[0]
		}

		result["scope"] = response.Scope
	}

	return result, nil
}
