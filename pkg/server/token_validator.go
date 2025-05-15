// Package server provides the MCP server implementation.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// KeyProvider is an interface for providing keys for JWT validation
type KeyProvider interface {
	// GetKey returns the key for validating a JWT token
	GetKey(token *jwt.Token) (interface{}, error)
}

// JWKSKeyProvider is a KeyProvider that fetches keys from a JWKS endpoint
type JWKSKeyProvider struct {
	jwksURL string
	// Add caching mechanism for keys
	keyCache map[string]interface{}
}

// NewJWKSKeyProvider creates a new JWKSKeyProvider
func NewJWKSKeyProvider(jwksURL string) *JWKSKeyProvider {
	return &JWKSKeyProvider{
		jwksURL:  jwksURL,
		keyCache: make(map[string]interface{}),
	}
}

// GetKey fetches the appropriate key from the JWKS endpoint
func (p *JWKSKeyProvider) GetKey(token *jwt.Token) (interface{}, error) {
	// Get the kid (key ID) from the token header
	kidInterface, ok := token.Header["kid"]
	if !ok {
		return nil, fmt.Errorf("token has no kid header")
	}

	kid, ok := kidInterface.(string)
	if !ok {
		return nil, fmt.Errorf("kid header is not a string")
	}

	// Check if key is in cache
	if key, ok := p.keyCache[kid]; ok {
		return key, nil
	}

	// Fetch the JWKS
	resp, err := http.Get(p.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	// Parse the JWKS
	var jwks struct {
		Keys []map[string]interface{} `json:"keys"`
	}

	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Find the key with matching kid
	for _, keyData := range jwks.Keys {
		if keyID, ok := keyData["kid"].(string); ok && keyID == kid {
			// Found the key, now parse it
			return parseJWK(keyData)
		}
	}

	return nil, fmt.Errorf("key with ID %s not found in JWKS", kid)
}

// parseJWK converts a JWK to a crypto key
func parseJWK(jwk map[string]interface{}) (interface{}, error) {
	// Only handling RSA keys in this example
	if kty, ok := jwk["kty"].(string); !ok || kty != "RSA" {
		return nil, fmt.Errorf("only RSA keys are supported, got %s", jwk["kty"])
	}

	// Process RSA key
	// In a real implementation, extract n (modulus) and e (exponent) and create an RSA public key

	// This is a simplified example - in a real implementation you would:
	// 1. Extract base64url-encoded n and e values
	// 2. Decode them
	// 3. Create an rsa.PublicKey with the values

	return nil, fmt.Errorf("JWK parsing not fully implemented")
}

// TokenIntrospector is an interface for token introspection (RFC 7662)
type TokenIntrospector interface {
	// IntrospectToken validates a token by calling an introspection endpoint
	IntrospectToken(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error)
}

// StandardIntrospector implements OAuth 2.0 token introspection
type StandardIntrospector struct {
	introspectionURL string
	clientID         string
	clientSecret     string
	httpClient       *http.Client
}

// NewStandardIntrospector creates a new StandardIntrospector
func NewStandardIntrospector(introspectionURL, clientID, clientSecret string) *StandardIntrospector {
	return &StandardIntrospector{
		introspectionURL: introspectionURL,
		clientID:         clientID,
		clientSecret:     clientSecret,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
	}
}

// IntrospectToken calls the token introspection endpoint to validate the token
func (i *StandardIntrospector) IntrospectToken(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error) {
	// Prepare the request
	data := url.Values{}
	data.Set("token", token)
	if tokenTypeHint != "" {
		data.Set("token_type_hint", tokenTypeHint)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", i.introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Set authentication if provided
	if i.clientID != "" && i.clientSecret != "" {
		req.SetBasicAuth(i.clientID, i.clientSecret)
	}

	// Make the request
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection failed with status code %d", resp.StatusCode)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	return result, nil
}

// JWTTokenValidator implements TokenValidator for JWT tokens
type JWTTokenValidator struct {
	config      *OAuthConfig
	keyProvider KeyProvider
}

// NewJWTTokenValidator creates a new JWTTokenValidator
func NewJWTTokenValidator(config *OAuthConfig, keyProvider KeyProvider) *JWTTokenValidator {
	return &JWTTokenValidator{
		config:      config,
		keyProvider: keyProvider,
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *JWTTokenValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, v.keyProvider.GetKey, jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("%w: invalid claims format", ErrInvalidToken)
	}

	// Validate required claims
	tokenClaims, err := v.validateClaims(claims)
	if err != nil {
		return nil, err
	}

	return tokenClaims, nil
}

// validateClaims validates the token claims
func (v *JWTTokenValidator) validateClaims(claims jwt.MapClaims) (*TokenClaims, error) {
	// Create result
	result := &TokenClaims{
		Claims: make(map[string]interface{}),
	}

	// Copy all claims to the result
	for k, v := range claims {
		result.Claims[k] = v
	}

	// Extract and validate subject
	if sub, ok := claims["sub"].(string); ok {
		result.Subject = sub
	} else {
		return nil, fmt.Errorf("%w: missing or invalid subject claim", ErrInvalidToken)
	}

	// Extract and validate issuer
	if iss, ok := claims["iss"].(string); ok {
		result.Issuer = iss
		// Validate issuer if configured
		if v.config.Issuer != "" && v.config.Issuer != iss {
			return nil, fmt.Errorf("%w: invalid issuer", ErrInvalidToken)
		}
	}

	// Extract and validate audience
	if aud, ok := claims["aud"].(string); ok {
		result.Audience = []string{aud}
		// Validate audience if configured
		if len(v.config.Audience) > 0 && !v.hasMatchingAudience([]string{aud}) {
			return nil, fmt.Errorf("%w: invalid audience", ErrInvalidToken)
		}
	} else if audList, ok := claims["aud"].([]interface{}); ok {
		// Handle audience as an array
		result.Audience = make([]string, 0, len(audList))
		for _, a := range audList {
			if audStr, ok := a.(string); ok {
				result.Audience = append(result.Audience, audStr)
			}
		}
		// Validate audience if configured
		if len(v.config.Audience) > 0 && !v.hasMatchingAudience(result.Audience) {
			return nil, fmt.Errorf("%w: invalid audience", ErrInvalidToken)
		}
	}

	// Extract expiration time
	if exp, ok := claims["exp"].(float64); ok {
		result.ExpiresAt = time.Unix(int64(exp), 0)
	}

	// Extract issued at time
	if iat, ok := claims["iat"].(float64); ok {
		result.IssuedAt = time.Unix(int64(iat), 0)
	}

	// Extract scopes
	if scope, ok := claims["scope"].(string); ok {
		result.Scopes = strings.Fields(scope)
	} else if scopes, ok := claims["scopes"].([]interface{}); ok {
		// Handle scopes as an array
		result.Scopes = make([]string, 0, len(scopes))
		for _, s := range scopes {
			if scopeStr, ok := s.(string); ok {
				result.Scopes = append(result.Scopes, scopeStr)
			}
		}
	}

	// Perform any required scope validation
	if len(v.config.RequiredScopes) > 0 {
		for _, requiredScope := range v.config.RequiredScopes {
			found := false
			for _, scope := range result.Scopes {
				if scope == requiredScope {
					found = true
					break
				}
			}
			if !found {
				return nil, ErrInsufficientScope
			}
		}
	}

	return result, nil
}

// hasMatchingAudience checks if the token audience matches any of the configured audiences
func (v *JWTTokenValidator) hasMatchingAudience(tokenAudiences []string) bool {
	for _, configAud := range v.config.Audience {
		for _, tokenAud := range tokenAudiences {
			if configAud == tokenAud {
				return true
			}
		}
	}
	return false
}

// IntrospectionTokenValidator implements TokenValidator using token introspection
type IntrospectionTokenValidator struct {
	config       *OAuthConfig
	introspector TokenIntrospector
}

// NewIntrospectionTokenValidator creates a new IntrospectionTokenValidator
func NewIntrospectionTokenValidator(config *OAuthConfig, introspector TokenIntrospector) *IntrospectionTokenValidator {
	return &IntrospectionTokenValidator{
		config:       config,
		introspector: introspector,
	}
}

// ValidateToken validates a token using introspection and returns the claims
func (v *IntrospectionTokenValidator) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	// Introspect the token
	result, err := v.introspector.IntrospectToken(ctx, token, "")
	if err != nil {
		return nil, fmt.Errorf("%w: introspection failed: %v", ErrInvalidToken, err)
	}

	// Check if token is active
	active, ok := result["active"].(bool)
	if !ok || !active {
		return nil, ErrInvalidToken
	}

	// Extract claims from introspection result
	claims := &TokenClaims{
		Claims: result,
	}

	// Extract subject
	if sub, ok := result["sub"].(string); ok {
		claims.Subject = sub
	} else {
		return nil, fmt.Errorf("%w: missing subject claim", ErrInvalidToken)
	}

	// Extract issuer
	if iss, ok := result["iss"].(string); ok {
		claims.Issuer = iss
		// Validate issuer if configured
		if v.config.Issuer != "" && v.config.Issuer != iss {
			return nil, fmt.Errorf("%w: invalid issuer", ErrInvalidToken)
		}
	}

	// Extract audience
	if aud, ok := result["aud"].(string); ok {
		claims.Audience = []string{aud}
		// Validate audience if configured
		if len(v.config.Audience) > 0 && !v.hasMatchingAudience([]string{aud}) {
			return nil, fmt.Errorf("%w: invalid audience", ErrInvalidToken)
		}
	} else if audList, ok := result["aud"].([]interface{}); ok {
		claims.Audience = make([]string, 0, len(audList))
		for _, a := range audList {
			if audStr, ok := a.(string); ok {
				claims.Audience = append(claims.Audience, audStr)
			}
		}
		// Validate audience if configured
		if len(v.config.Audience) > 0 && !v.hasMatchingAudience(claims.Audience) {
			return nil, fmt.Errorf("%w: invalid audience", ErrInvalidToken)
		}
	}

	// Extract expiration time
	if exp, ok := result["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}

	// Extract issued at time
	if iat, ok := result["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}

	// Extract scopes
	if scope, ok := result["scope"].(string); ok {
		claims.Scopes = strings.Fields(scope)
	} else if scopes, ok := result["scopes"].([]interface{}); ok {
		claims.Scopes = make([]string, 0, len(scopes))
		for _, s := range scopes {
			if scopeStr, ok := s.(string); ok {
				claims.Scopes = append(claims.Scopes, scopeStr)
			}
		}
	}

	// Perform any required scope validation
	if len(v.config.RequiredScopes) > 0 {
		for _, requiredScope := range v.config.RequiredScopes {
			found := false
			for _, scope := range claims.Scopes {
				if scope == requiredScope {
					found = true
					break
				}
			}
			if !found {
				return nil, ErrInsufficientScope
			}
		}
	}

	return claims, nil
}

// hasMatchingAudience checks if the token audience matches any of the configured audiences
func (v *IntrospectionTokenValidator) hasMatchingAudience(tokenAudiences []string) bool {
	for _, configAud := range v.config.Audience {
		for _, tokenAud := range tokenAudiences {
			if configAud == tokenAud {
				return true
			}
		}
	}
	return false
}

// CompositeTokenValidator implements TokenValidator by trying multiple validators in sequence
type CompositeTokenValidator struct {
	validators []TokenValidator
}

// NewCompositeTokenValidator creates a new CompositeTokenValidator
func NewCompositeTokenValidator(validators ...TokenValidator) *CompositeTokenValidator {
	return &CompositeTokenValidator{
		validators: validators,
	}
}

// ValidateToken validates a token using all configured validators until one succeeds
func (v *CompositeTokenValidator) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	var lastError error

	for _, validator := range v.validators {
		claims, err := validator.ValidateToken(ctx, token)
		if err == nil {
			return claims, nil
		}
		lastError = err
	}

	if lastError != nil {
		return nil, lastError
	}

	return nil, errors.New("no validators configured")
}
