package mware

import (
	"net/http"

	"zgo.at/zhttp"
)

// WrapWriter replaces the http.ResponseWriter with our version of
// http.ResponseWriter for some additional functionality.
func WrapWriter() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(zhttp.NewResponseWriter(w, r.ProtoMajor), r)
		})
	}

}
