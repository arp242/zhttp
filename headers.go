package zhttp

import "net/http"

func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		// TODO: Don't load gfonts, use proper cfg.Static, get rid of
		// unsafe-inline
		//w.Header().Set("Content-Security-Policy", "default-src localhost:8081; style-src localhost:8081 https://fonts.googleapis.com 'unsafe-inline'; font-src localhost:8081 https://fonts.gstatic.com;")
		w.Header().Set("Strict-Transport-Security", "max-age=2592000")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		next.ServeHTTP(w, r)
	})
}
