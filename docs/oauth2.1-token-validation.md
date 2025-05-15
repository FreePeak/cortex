# OAuth 2.1 Token Validation Implementation

This document describes the token validation mechanisms implemented for the Model Context Protocol (MCP) 2025-03-26 specification.

## Overview

The OAuth 2.1 token validation system in Cortex supports multiple token formats and validation methods:

1. JWT Token Validation - For validating JSON Web Tokens with digital signatures
2. Token Introspection - For validating opaque tokens against an authorization server
3. Composite Validation - For trying multiple validation methods in sequence

## JWT Token Validation

JWT (JSON Web Token) validation verifies tokens that contain their own claims and are signed with a cryptographic key. 

### Features

- Validation of standard JWT claims (issuer, audience, expiration, etc.)
- Support for scope validation (both space-delimited strings and arrays)
- RSA signature verification (RS256, RS384, RS512)
- JSON Web Key Set (JWKS) support for fetching public keys

### Key Components

#### JWTTokenValidator

The primary component that validates JWT tokens:

```go
type JWTTokenValidator struct {
    config      *OAuthConfig
    keyProvider KeyProvider
}
```

- `config` - Configuration settings for validation (issuer, audience, etc.)
- `keyProvider` - Source of cryptographic keys for signature verification

#### KeyProvider Interface

An interface for retrieving the key needed to validate a token:

```go
type KeyProvider interface {
    GetKey(token *jwt.Token) (interface{}, error)
}
```

#### JWKSKeyProvider

A KeyProvider implementation that fetches keys from a JWKS endpoint:

```go
type JWKSKeyProvider struct {
    jwksURL  string
    keyCache map[string]interface{}
}
```

### Validation Process

1. Parse the JWT token
2. Verify the token's signature using the KeyProvider
3. Extract and validate claims (subject, issuer, audience, etc.)
4. Check if the token has all required scopes
5. Return the validated TokenClaims

## Token Introspection

Token introspection validates opaque tokens by sending them to an authorization server, following RFC 7662.

### Features

- Support for the OAuth 2.0 Token Introspection protocol
- Validation of tokens against an authorization server
- Client authentication for introspection endpoints
- Mapping of introspection responses to TokenClaims

### Key Components

#### IntrospectionTokenValidator

The primary component that validates tokens via introspection:

```go
type IntrospectionTokenValidator struct {
    config       *OAuthConfig
    introspector TokenIntrospector
}
```

- `config` - Configuration settings for validation
- `introspector` - Component that performs the actual introspection request

#### TokenIntrospector Interface

An interface for sending token introspection requests:

```go
type TokenIntrospector interface {
    IntrospectToken(ctx context.Context, token, tokenTypeHint string) (map[string]interface{}, error)
}
```

#### StandardIntrospector

A standard implementation of the TokenIntrospector interface:

```go
type StandardIntrospector struct {
    introspectionURL string
    clientID         string
    clientSecret     string
    httpClient       *http.Client
}
```

### Validation Process

1. Send the token to the introspection endpoint
2. Check if the token is active
3. Extract claims from the introspection response
4. Validate issuer and audience if configured
5. Check if the token has all required scopes
6. Return the validated TokenClaims

## Composite Validation

The CompositeTokenValidator allows trying multiple validators in sequence, which is useful when supporting multiple token formats.

```go
type CompositeTokenValidator struct {
    validators []TokenValidator
}
```

It tries each validator in sequence until one succeeds or all fail.

## Usage Example

```go
// Create a JWKS key provider
keyProvider := NewJWKSKeyProvider("https://auth-server.example.com/.well-known/jwks.json")

// Create a JWT validator
jwtValidator := NewJWTTokenValidator(&OAuthConfig{
    Issuer:         "https://auth-server.example.com",
    Audience:       []string{"api-client"},
    RequiredScopes: []string{"tools:read"},
}, keyProvider)

// Create a token introspection validator
introspector := NewStandardIntrospector(
    "https://auth-server.example.com/oauth/introspect",
    "client-id",
    "client-secret",
)
introspectionValidator := NewIntrospectionTokenValidator(&OAuthConfig{
    Issuer:         "https://auth-server.example.com",
    Audience:       []string{"api-client"},
    RequiredScopes: []string{"tools:read"},
}, introspector)

// Create a composite validator that tries both
validator := NewCompositeTokenValidator(jwtValidator, introspectionValidator)

// Use the validator in middleware
middleware := NewOAuthMiddleware(validator)
```

## Security Considerations

1. **Key Management**: Keys used for JWT validation should be rotated regularly
2. **Token Lifetime**: JWT tokens should have short lifetimes due to their lack of revocation
3. **Introspection Caching**: Consider caching introspection results to reduce load on the authorization server
4. **Transport Security**: All communication with authorization servers must use TLS

## Next Steps

The token validation implementation provides the foundation for securing the MCP server. The next steps involve:

1. Implementing middleware for token validation
2. Adding scope-based permission checks for tools
3. Integrating with PocketBase authentication
4. Creating documentation for authorization setup 