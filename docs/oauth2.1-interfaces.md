# OAuth 2.1 Authentication Interfaces

This document describes the OAuth 2.1 authentication interfaces implemented in Cortex for MCP 2025-03-26 alignment.

## Overview

The authentication framework implements OAuth 2.1 standards to provide a secure and flexible way to authenticate clients and authorize access to tools and resources. The interfaces are designed to be extensible and work with various OAuth providers.

## Core Interfaces

### TokenValidator

This interface is responsible for validating OAuth 2.1 tokens:

```go
type TokenValidator interface {
    ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}
```

- **ValidateToken**: Validates the provided token and returns the claims if valid.

### ScopeChecker

This interface provides methods to check token scopes against required permissions:

```go
type ScopeChecker interface {
    HasScope(claims *TokenClaims, requiredScope string) bool
    HasAnyScope(claims *TokenClaims, requiredScopes ...string) bool
    HasAllScopes(claims *TokenClaims, requiredScopes ...string) bool
}
```

- **HasScope**: Checks if the token has the required scope.
- **HasAnyScope**: Checks if the token has any of the required scopes.
- **HasAllScopes**: Checks if the token has all the required scopes.

## Types and Structures

### TokenClaims

This structure represents the validated claims from an OAuth 2.1 access token:

```go
type TokenClaims struct {
    Subject   string
    Issuer    string
    Audience  []string
    ExpiresAt time.Time
    IssuedAt  time.Time
    Scopes    []string
    Claims    map[string]interface{}
}
```

### OAuthMiddleware

This structure provides middleware for OAuth 2.1 authentication:

```go
type OAuthMiddleware struct {
    validator TokenValidator
    checker   ScopeChecker
}
```

#### Key Methods:

- **NewOAuthMiddleware**: Creates a new OAuthMiddleware with the provided token validator.
- **Middleware**: Returns an http.Handler middleware that validates OAuth tokens.
- **RequireScope**: Returns middleware that ensures the token has the required scope.
- **RequireAnyScope**: Returns middleware that ensures the token has at least one of the required scopes.
- **RequireAllScopes**: Returns middleware that ensures the token has all of the required scopes.

### OAuthConfig

Configuration options for OAuth 2.1 authorization:

```go
type OAuthConfig struct {
    Issuer            string
    Audience          []string
    JWKSUrl           string
    RequiredScopes    []string
    TokenLookupScheme string
    TokenHeaderName   string
    TokenQueryParam   string
}
```

## Error Types

The following errors are defined for OAuth 2.1 authentication:

- **ErrInvalidToken**: Indicates that the provided token is invalid or expired.
- **ErrInsufficientScope**: Indicates that the token does not have the required scope.
- **ErrMissingToken**: Indicates that no token was provided.
- **ErrInvalidRequest**: Indicates an invalid OAuth request.

## Helper Functions

- **GetTokenClaimsFromContext**: Extracts token claims from the request context.
- **DefaultOAuthConfig**: Returns a default configuration for OAuth 2.1.

## Usage Example

Here's a simple example of how to use these interfaces to protect API endpoints:

```go
// Create a token validator implementation
validator := &MyTokenValidator{} 

// Create OAuth middleware
middleware := NewOAuthMiddleware(validator)

// Apply middleware to routes
router.Use(middleware.Middleware)

// Apply scope-specific middleware to protected routes
router.Handle("/tools", middleware.RequireScope("tools:read", toolsHandler))
router.Handle("/admin", middleware.RequireAllScopes([]string{"admin", "tools:manage"}, adminHandler))
```

## Next Steps

The interfaces defined here provide the foundation for OAuth 2.1 authentication in Cortex. The next tasks involve:

1. Implementing a concrete TokenValidator using JWT validation
2. Implementing token introspection for opaque tokens
3. Adding scope-based permission system for tools
4. Integrating with PocketBase's authentication system

## References

- [OAuth 2.1 Specification](https://oauth.net/2.1/)
- [JWT (JSON Web Tokens)](https://jwt.io)
- [OAuth 2.0 Token Introspection](https://www.rfc-editor.org/rfc/rfc7662) 