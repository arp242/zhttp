package mware

import (
	"net/http"

	"zgo.at/zhttp"
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
func Headers(h http.Header) zhttp.Middleware {
	headers := make(http.Header)
	for k, v := range DefaultHeaders {
		headers[http.CanonicalHeaderKey(k)] = v
	}
	for k, v := range h {
		headers[http.CanonicalHeaderKey(k)] = v
	}

	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			for k, v := range headers {
				for _, v2 := range v {
					w.Header().Add(k, v2)
				}
			}

			return next(w, r)
		})
	}
}
