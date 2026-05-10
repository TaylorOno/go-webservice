package web

import (
	"context"
	"net/http"
	"strings"
)

// PlatformHeaders is used to identify what headers are to be added to the request context
// which is used by PassToContext middleware
type PlatformHeaders struct {
	ContextKey        interface{}
	StartsWithHeaders []string
	ExactHeaders      []string
}

type HeaderOption func(*PlatformHeaders)

func StartsWithHeaders(headers []string) HeaderOption {
	return func(h *PlatformHeaders) {
		h.StartsWithHeaders = make([]string, len(headers))
		for i, v := range headers {
			h.StartsWithHeaders[i] = http.CanonicalHeaderKey(v)
		}
	}
}

func ExactHeaders(headers []string) HeaderOption {
	return func(h *PlatformHeaders) {
		h.ExactHeaders = make([]string, len(headers))
		for i, v := range headers {
			h.ExactHeaders[i] = http.CanonicalHeaderKey(v)
		}
	}
}

// NewPlatformHeaders creates a new PlatformHeaders middleware provider for a given context key.
func NewPlatformHeaders(contextKey interface{}, opts ...HeaderOption) *PlatformHeaders {
	ph := &PlatformHeaders{ContextKey: contextKey}
	for _, opt := range opts {
		opt(ph)
	}

	return ph
}

// AddToContext is middleware to add the headers to the request context.
func (h *PlatformHeaders) AddToContext(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(h.headersToContext(r)))
	}

	return fn
}

// headersToContext adds header information to the request context
func (h *PlatformHeaders) headersToContext(r *http.Request) context.Context {
	ctx := r.Context()
	httpHeaders := http.Header{}

HeaderLoop:
	for name, value := range r.Header {
		canonicalName := http.CanonicalHeaderKey(name)
		for _, exact := range h.ExactHeaders {
			exact = http.CanonicalHeaderKey(exact)
			if canonicalName == exact {
				vCopy := make([]string, len(value))
				copy(vCopy, value)
				httpHeaders[canonicalName] = vCopy
				continue HeaderLoop
			}
		}

		for _, start := range h.StartsWithHeaders {
			start = http.CanonicalHeaderKey(start)
			if strings.HasPrefix(canonicalName, start) {
				vCopy := make([]string, len(value))
				copy(vCopy, value)
				httpHeaders[canonicalName] = vCopy
				continue HeaderLoop
			}
		}
	}

	return context.WithValue(ctx, h.ContextKey, httpHeaders)
}
