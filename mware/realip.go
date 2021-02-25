package mware

import (
	"net/http"
	"strings"

	"zgo.at/zstd/znet"
)

// RealIP sets the RemoteAddr to X-Real-Ip, X-Forwarded-For, or the RemoteAddr
// without a port.
//
// The end result willl never have a source port set.
func RealIP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.RemoteAddr = realIP(r)
			next.ServeHTTP(w, r)
		})
	}
}

func realIP(r *http.Request) string {
	cfip := r.Header.Get("Cf-Connecting-Ip")
	if cfip != "" && !znet.PrivateIPString(cfip) {
		return cfip
	}

	xrip := r.Header.Get("X-Real-Ip")
	if xrip != "" && !znet.PrivateIPString(xrip) {
		return xrip
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		xffSplit := strings.Split(xff, ",")
		for i := len(xffSplit) - 1; i >= 0; i-- {
			if !znet.PrivateIPString(xffSplit[i]) {
				return strings.TrimSpace(xffSplit[i])
			}
		}
	}

	return znet.RemovePort(r.RemoteAddr)
}
