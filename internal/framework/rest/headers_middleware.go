package rest

import (
	"net/http"
)

// AddHeadersFromContext is used to add headers that were passed via request context
func AddHeadersFromContext(contextKey interface{}) ClientMiddleware {
	return func(c Doer) Doer {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			ctxHeaders, ok := r.Context().Value(contextKey).(http.Header)
			if ok {
				for key := range ctxHeaders {
					r.Header.Set(key, ctxHeaders.Get(key))
				}
			}

			return c.Do(r)
		})
	}
}
