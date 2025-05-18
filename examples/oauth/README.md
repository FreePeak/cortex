# OAuth 2.1 Examples for Cortex

This directory contains examples demonstrating how to implement OAuth 2.1 authentication with Cortex servers, as part of the Model Context Protocol (MCP) 2025-03-26 specification alignment.

## Structure

The examples are organized into the following directories:

- `minimal/`: A minimal implementation focusing on OAuth 2.1 authentication only
- `server/`: A complete server implementation with OAuth 2.1 and tool permissions

## Running the Examples

### Minimal OAuth Example

```bash
cd minimal
go run main.go
```

This starts a simple HTTP server on port 8080 with the following endpoints:

- `/protected`: A basic protected endpoint requiring a valid OAuth token
- `/admin`: An endpoint requiring the "cortex:admin" scope
- `/high-privilege`: An endpoint requiring multiple scopes

### Full Server Example

```bash
cd server
go run main.go
```

This starts a Cortex MCP server with the following endpoints:

- `/`: A public welcome page (no authentication required)
- `/api/mcp/`: The main MCP server endpoint protected by OAuth 2.1
- `/admin`: An admin endpoint requiring the "cortex:admin" scope
- `/tools/echo`: An endpoint for the echo tool requiring tool-specific permissions

## Testing the OAuth Examples

For testing purposes, both examples include a mock token validator that accepts "test-token" as a valid token.

To test the protected endpoints, you can use curl:

```bash
# Access a protected endpoint with a test token
curl -H "Authorization: Bearer test-token" http://localhost:8080/protected

# Access an endpoint requiring specific scopes
curl -H "Authorization: Bearer test-token" http://localhost:8080/admin
```

## Using in Production

For production use, replace the mock validator with a real JWT validator:

1. Configure your OAuth 2.1 provider (Auth0, Keycloak, etc.)
2. Update the OAuth configuration with your provider's information (issuer, audience, JWKS URL)
3. Use the `JWTTokenValidator` with a `JWKSKeyProvider` to validate tokens

For opaque tokens, use the `IntrospectionTokenValidator` with your token introspection endpoint.

## Key Components

- **OAuthConfig**: Configuration for OAuth 2.1 settings
- **TokenValidator**: Interface for validating tokens
- **OAuthMiddleware**: HTTP middleware for OAuth token validation
- **TokenClaims**: Structure representing the validated claims from a token
- **ToolPermissions**: Manager for tool-specific permissions

## Scope Conventions

The examples use the following scope naming conventions:

- `cortex:api`: General API access
- `cortex:admin`: Administrative access
- `cortex:tool:read`: Read access to all tools
- `cortex:tool:execute:X`: Execute permission for a specific tool X 