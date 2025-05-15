# OAuth 2.1 Authorization Setup and Configuration

This document explains how to set up and configure OAuth 2.1 authorization in Cortex. The authorization framework provides secure access control to Cortex API endpoints and tools, following OAuth 2.1 standards.

## Overview

Cortex implements a comprehensive OAuth 2.1 authorization framework with the following features:

- JWT token validation
- Scope-based access control
- Tool-specific permissions
- Multiple token extraction methods (header, query parameter, cookie)
- Integration with external identity providers

## Basic Setup

### Step 1: Configure OAuth Settings

Create an `OAuthConfig` with your authorization settings:

```go
config := &server.OAuthConfig{
    Issuer:            "https://auth.example.com",     // Your OAuth issuer URL
    Audience:          []string{"cortex-api"},         // Expected audience values
    JWKSUrl:           "https://auth.example.com/.well-known/jwks.json", // JWKS endpoint
    RequiredScopes:    []string{"cortex:api"},         // Global required scopes
    TokenLookupScheme: "header,query",                 // Where to look for tokens
    TokenHeaderName:   "Authorization",                // Header name (default)
    TokenQueryParam:   "access_token",                 // Query parameter name
}
```

### Step 2: Create a Token Validator

Choose one of these validator implementations based on your needs:

#### JWT Token Validator (recommended)

```go
// Create a key provider that fetches keys from your JWKS endpoint
keyProvider := server.NewJWKSKeyProvider(config.JWKSUrl)

// Create a JWT validator
validator := server.NewJWTTokenValidator(config, keyProvider)

// Create OAuth middleware
middleware := server.NewOAuthMiddlewareWithConfig(validator, config)
```

#### OAuth 2.0 Introspection Validator

```go
// Create an introspector for RFC 7662 token introspection
introspector := server.NewStandardIntrospector(
    "https://auth.example.com/oauth/introspect", 
    "client_id", 
    "client_secret"
)

// Create an introspection validator
validator := server.NewIntrospectionTokenValidator(config, introspector)

// Create OAuth middleware
middleware := server.NewOAuthMiddlewareWithConfig(validator, config)
```

### Step 3: Apply OAuth Middleware

Apply the OAuth middleware to your HTTP handlers:

```go
// Create your handler
handler := http.HandlerFunc(yourHandlerFunc)

// Wrap with OAuth middleware
protectedHandler := middleware.Middleware(handler)

// Use in your HTTP server
http.Handle("/api/protected", protectedHandler)
```

## Scope-Based Access Control

### Defining Scopes

Scopes are strings that represent permissions. In Cortex, we use a hierarchical naming convention:

- `cortex:api` - General API access
- `cortex:tool:read` - Read access to all tools
- `cortex:tool:execute:{toolName}` - Execute permission for a specific tool

### Requiring Scopes for Endpoints

You can protect endpoints with scope requirements:

```go
// Require a single scope
handler := middleware.RequireScope("cortex:api", yourHandler)

// Require any one of multiple scopes
handler := middleware.RequireAnyScope([]string{"cortex:admin", "cortex:tool:read"}, yourHandler)

// Require all specified scopes
handler := middleware.RequireAllScopes([]string{"cortex:api", "cortex:tool:read"}, yourHandler)
```

## Tool Permissions

Cortex provides a dedicated permission system for tools with three permission types:

- `ToolPermissionRead`: Read tool metadata
- `ToolPermissionWrite`: Modify tool configuration
- `ToolPermissionExecute`: Execute the tool

### Setting Up Tool Permissions

```go
// Create tool permissions with the OAuth middleware
toolPermissions := server.NewToolPermissions(middleware)

// Protect a tool endpoint
handler := toolPermissions.RequireToolPermission(
    "calculator",                // Tool name
    server.ToolPermissionExecute, // Permission type
    yourToolHandler              // Handler to protect
)
```

### Tool Permission Scopes

Tool permissions use the following scope format:

- `cortex:tool:{permission}:{toolName}`

Examples:
- `cortex:tool:read` - Global read access to all tools
- `cortex:tool:execute:calculator` - Execute permission for the calculator tool
- `cortex:tool:write:weather` - Write permission for the weather tool

## PocketBase Integration

If you're using PocketBase, you can set up OAuth with the Cortex plugin:

```go
// Create the plugin
plugin := pocketbase.NewCortexPlugin()

// Set up OAuth
validator := CreateYourValidator() // See validator setup above
middleware := server.NewOAuthMiddlewareWithConfig(validator, config)

// Add OAuth to the plugin
plugin.WithOAuth(middleware).WithOAuthConfig(config)

// Register with PocketBase
plugin.RegisterWithPocketBase(app)
```

## Accessing Token Claims

In your HTTP handlers, you can access token claims from the context:

```go
func yourHandler(w http.ResponseWriter, r *http.Request) {
    // Get token claims from context
    claims, ok := server.GetTokenClaimsFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Use claims information
    userID := claims.Subject
    scopes := claims.Scopes
    
    // Check permissions manually if needed
    if !hasPermission(claims) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // Continue with authorized operation...
}
```

## Custom Scope Checking

If you need custom scope checking logic, you can implement the `ScopeChecker` interface:

```go
type CustomScopeChecker struct {
    // Any fields you need
}

func (c *CustomScopeChecker) HasScope(claims *server.TokenClaims, requiredScope string) bool {
    // Your custom logic here
    return customScopeCheckLogic(claims, requiredScope)
}

func (c *CustomScopeChecker) HasAnyScope(claims *server.TokenClaims, requiredScopes ...string) bool {
    // Your custom logic here
    return customAnyScopeCheckLogic(claims, requiredScopes)
}

func (c *CustomScopeChecker) HasAllScopes(claims *server.TokenClaims, requiredScopes ...string) bool {
    // Your custom logic here
    return customAllScopesCheckLogic(claims, requiredScopes)
}

// Then use your custom checker:
middleware.WithScopeChecker(&CustomScopeChecker{})
```

## Security Considerations

1. **Token Validation**: Always validate tokens for integrity, expiration, issuer, and audience.
2. **HTTPS**: Use HTTPS for all communication to protect tokens.
3. **Proper Scopes**: Grant minimal necessary scopes to each client.
4. **Token Storage**: Securely store tokens client-side, and never in localStorage.
5. **Token Expiration**: Use short-lived access tokens with refresh token rotation.
6. **CORS**: Configure CORS properly to restrict access to trusted domains.

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Indicates invalid or expired token, or missing token.
2. **403 Forbidden**: Valid token but insufficient scopes for the requested action.
3. **JWKS Key Issues**: If the key ID (kid) in the token doesn't match any key in the JWKS.

### Debugging Tips

- Use the `Authorization` header debug logs for token extraction issues.
- Check token expiration and issuer if validation fails.
- Verify the token has the proper scopes for the requested action.
- Ensure your JWKS endpoint is accessible and returns the correct keys.

## Example Configuration

Here's a complete example of setting up OAuth 2.1 with JWT validation:

```go
func SetupOAuth() *server.OAuthMiddleware {
    // Create OAuth configuration
    config := &server.OAuthConfig{
        Issuer:            "https://auth.example.com",
        Audience:          []string{"cortex-api"},
        JWKSUrl:           "https://auth.example.com/.well-known/jwks.json",
        RequiredScopes:    []string{"cortex:api"},
        TokenLookupScheme: "header,query,cookie",
        TokenHeaderName:   "Authorization",
        TokenQueryParam:   "access_token",
    }
    
    // Create key provider
    keyProvider := server.NewJWKSKeyProvider(config.JWKSUrl)
    
    // Create JWT validator
    validator := server.NewJWTTokenValidator(config, keyProvider)
    
    // Create and return OAuth middleware
    return server.NewOAuthMiddlewareWithConfig(validator, config)
}
``` 