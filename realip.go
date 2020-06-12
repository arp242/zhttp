package zhttp

import (
	"net/http"
	"strings"
)

// RealIP sets the RemoteAddr to X-Real-Ip, X-Forwarded-For, or the RemoteAddr
// without a port.
func RealIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = realIP(r)
		h.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request) string {
	xrip := r.Header.Get("X-Real-Ip")
	if xrip != "" {
		return xrip
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		return xff[:i]
	}

	return RemovePort(r.RemoteAddr)
}
