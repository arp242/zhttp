package mware

import (
	"net/http"
)

// DefaultHeaders will be set by default.
var DefaultHeaders = http.Header{
	"Strict-Transport-Security": []string{"max-age=7776000"},
	"X-Frame-Options":           []string{"deny"},
	"X-Content-Type-Options":    []string{"nosniff"},
}

// Headers sets the given headers.
//
// DefaultHeaders will always be set. Headers passed to this function overrides
// them. Use a nil value to remove a header.
func Headers(h http.Header) func(next http.Handler) http.Handler {
	headers := make(http.Header)
	for k, v := range DefaultHeaders {
		headers[http.CanonicalHeaderKey(k)] = v
	}
	for k, v := range h {
		headers[http.CanonicalHeaderKey(k)] = v
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range headers {
				for _, v2 := range v {
					w.Header().Add(k, v2)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
