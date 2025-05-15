package server

import (
	"fmt"
	"net/http"
)

// ToolPermission represents the type of permission required for a tool
type ToolPermission string

const (
	// ToolPermissionRead allows reading tool metadata
	ToolPermissionRead ToolPermission = "read"

	// ToolPermissionWrite allows modifying tool configuration
	ToolPermissionWrite ToolPermission = "write"

	// ToolPermissionExecute allows executing the tool
	ToolPermissionExecute ToolPermission = "execute"
)

// ToolPermissions provides scope-based tool access control
type ToolPermissions struct {
	oauth *OAuthMiddleware
}

// NewToolPermissions creates a new tool permissions system
func NewToolPermissions(oauth *OAuthMiddleware) *ToolPermissions {
	return &ToolPermissions{
		oauth: oauth,
	}
}

// FormatToolScope formats a tool permission scope string
// Format: cortex:tool:[permission]:[toolName]
// If toolName is empty, a global permission scope is returned
func (tp *ToolPermissions) FormatToolScope(toolName string, permission ToolPermission) string {
	if toolName == "" {
		return fmt.Sprintf("cortex:tool:%s", permission)
	}
	return fmt.Sprintf("cortex:tool:%s:%s", permission, toolName)
}

// RequireToolPermission returns middleware that checks if the token has the required tool permission
func (tp *ToolPermissions) RequireToolPermission(toolName string, permission ToolPermission, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token claims from context
		claims, ok := GetTokenClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized: No token claims found", http.StatusUnauthorized)
			return
		}

		// Check for specific tool permission
		specificScope := tp.FormatToolScope(toolName, permission)

		// Check for global permission (applies to read only)
		globalScope := ""
		if permission == ToolPermissionRead {
			globalScope = tp.FormatToolScope("", permission)
		}

		// Check if token has required scopes
		hasPermission := false

		// First check for specific tool permission
		for _, scope := range claims.Scopes {
			if scope == specificScope {
				hasPermission = true
				break
			}
		}

		// If not found and global scope is applicable, check for global permission
		if !hasPermission && globalScope != "" {
			for _, scope := range claims.Scopes {
				if scope == globalScope {
					hasPermission = true
					break
				}
			}
		}

		if !hasPermission {
			http.Error(
				w,
				fmt.Sprintf("Forbidden: Insufficient scope, requires %s", specificScope),
				http.StatusForbidden,
			)
			return
		}

		// Proceed to next handler
		next.ServeHTTP(w, r)
	})
}

// HasToolPermission checks if the given claims have permission for a specific tool
func (tp *ToolPermissions) HasToolPermission(claims *TokenClaims, toolName string, permission ToolPermission) bool {
	specificScope := tp.FormatToolScope(toolName, permission)

	// Check for global permission (applies to read only)
	globalScope := ""
	if permission == ToolPermissionRead {
		globalScope = tp.FormatToolScope("", permission)
	}

	// Check specific permission
	for _, scope := range claims.Scopes {
		if scope == specificScope {
			return true
		}
	}

	// Check global permission if applicable
	if globalScope != "" {
		for _, scope := range claims.Scopes {
			if scope == globalScope {
				return true
			}
		}
	}

	return false
}

// HasAnyToolPermission checks if the token has any permission for the given tool
func (tp *ToolPermissions) HasAnyToolPermission(claims *TokenClaims, toolName string) bool {
	return tp.HasToolPermission(claims, toolName, ToolPermissionRead) ||
		tp.HasToolPermission(claims, toolName, ToolPermissionWrite) ||
		tp.HasToolPermission(claims, toolName, ToolPermissionExecute)
}
