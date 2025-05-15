package server

import (
	"net/http"
	"strings"
)

// tokenExtractor is a function that extracts a token from an HTTP request
type tokenExtractor func(r *http.Request) string

// NewOAuthMiddlewareWithConfig creates a new OAuthMiddleware with the provided configuration
func NewOAuthMiddlewareWithConfig(validator TokenValidator, config *OAuthConfig) *OAuthMiddleware {
	if config == nil {
		config = DefaultOAuthConfig()
	}

	middleware := &OAuthMiddleware{
		validator:       validator,
		checker:         &defaultScopeChecker{},
		config:          config,
		tokenExtractors: []tokenExtractor{},
	}

	// Set up token extractors based on configuration
	middleware.setupTokenExtractors()

	return middleware
}

// setupTokenExtractors configures the token extractors based on the middleware configuration
func (m *OAuthMiddleware) setupTokenExtractors() {
	m.tokenExtractors = []tokenExtractor{}

	// Parse the lookup scheme to determine how to extract tokens
	schemes := strings.Split(m.config.TokenLookupScheme, ",")
	for _, scheme := range schemes {
		scheme = strings.TrimSpace(scheme)
		switch scheme {
		case "header":
			headerName := m.config.TokenHeaderName
			if headerName == "" {
				headerName = "Authorization"
			}
			m.tokenExtractors = append(m.tokenExtractors, func(r *http.Request) string {
				return extractTokenFromHeader(r, headerName)
			})
		case "query":
			paramName := m.config.TokenQueryParam
			if paramName == "" {
				paramName = "token"
			}
			m.tokenExtractors = append(m.tokenExtractors, func(r *http.Request) string {
				return r.URL.Query().Get(paramName)
			})
		case "cookie":
			cookieName := m.config.TokenQueryParam
			if cookieName == "" {
				cookieName = "token"
			}
			m.tokenExtractors = append(m.tokenExtractors, func(r *http.Request) string {
				cookie, err := r.Cookie(cookieName)
				if err != nil {
					return ""
				}
				return cookie.Value
			})
		}
	}

	// If no extractors were configured, add the default one (Authorization header)
	if len(m.tokenExtractors) == 0 {
		m.tokenExtractors = append(m.tokenExtractors, func(r *http.Request) string {
			return extractTokenFromHeader(r, "Authorization")
		})
	}
}

// extractTokenFromHeader extracts a token from the given header
func extractTokenFromHeader(r *http.Request, headerName string) string {
	// Special handling for Authorization header (expect "Bearer token")
	if headerName == "Authorization" {
		authHeader := r.Header.Get(headerName)
		if authHeader == "" {
			return ""
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return ""
		}

		return parts[1]
	}

	// For other headers, just return the value
	return r.Header.Get(headerName)
}

// extractToken tries all configured token extractors in order until it finds a token
func (m *OAuthMiddleware) extractToken(r *http.Request) string {
	for _, extractor := range m.tokenExtractors {
		if token := extractor(r); token != "" {
			return token
		}
	}
	return ""
}
