package server

import (
	"net/http"
	"net/url"
)

// HTTPAdapterOption is a function that configures an HTTP adapter.
type HTTPAdapterOption func(*HTTPAdapter)

// WithPath sets the base path for the HTTP adapter.
func WithPath(basePath string) HTTPAdapterOption {
	return func(a *HTTPAdapter) {
		a.basePath = basePath
	}
}

// HTTPAdapter adapts an Embeddable to be used in an HTTP server.
type HTTPAdapter struct {
	embeddable Embeddable
	basePath   string
}

// NewHTTPAdapter creates a new HTTPAdapter.
func NewHTTPAdapter(embeddable Embeddable, opts ...HTTPAdapterOption) *HTTPAdapter {
	adapter := &HTTPAdapter{
		embeddable: embeddable,
		basePath:   "/mcp",
	}

	// Apply options
	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// Handler returns an http.Handler that can be registered with an HTTP server.
func (a *HTTPAdapter) Handler() http.Handler {
	// Create a router that handles MCP routes
	baseHandler := a.embeddable.ToHTTPHandler()

	// Return a handler that checks the path and delegates to the MCP handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path starts with the MCP base path
		if len(r.URL.Path) >= len(a.basePath) && r.URL.Path[:len(a.basePath)] == a.basePath {
			// Strip the base path from the request
			r2 := new(http.Request)
			*r2 = *r

			// Remove the base path
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = r.URL.Path[len(a.basePath):]
			if r2.URL.Path == "" {
				r2.URL.Path = "/"
			}

			// Delegate to the MCP handler
			baseHandler.ServeHTTP(w, r2)
			return
		}

		// Not an MCP request, return 404
		http.NotFound(w, r)
	})
}
