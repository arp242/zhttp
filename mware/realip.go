package mware

import (
	"net/http"
	"strings"

	"zgo.at/zstd/znet"
)

// RealIP sets the RemoteAddr to CF-Connecting-IP, Fly-Client-IP,
// X-Azure-SocketIP, X-Real-IP, X-Forwarded-For, or the RemoteAddr without a
// port.
//
// The end result willl never have a source port set. It will ignore local and
// private addresses such as 127.0.0.1, 192.168.1.1, etc.
//
// TODO: allow configuring which headers to look at, as this is very much
// dependent on the specific configuration; preferring e.g. Fly-Client-Ip means
// its trivial to spoof the "real IP".
func RealIP(never ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.RemoteAddr = realIP(r)
			next.ServeHTTP(w, r)
		})
	}
}

// See: https://github.com/arp242/zhttp/pull/5
func realIP(r *http.Request) string {
	// Prefer the "real IP" as reported by proxies such as nginx, Cloudflare,
	// Akamai, Fly, etc. as that's usually more accurate.
	for _, h := range []string{"Cf-Connecting-Ip", "Fly-Client-Ip", "X-Azure-Socketip", "X-Real-Ip"} {
		if ip := r.Header.Get(h); ip != "" && !znet.PrivateIPString(ip) {
			return ip
		}
	}

	// Every proxy appends a new value; so the left-most is the closest to the
	// client, and the right-most is whatever is proxying to us. We use the
	// right-most, because in *most* cases that's what you care about: "who is
	// connecting to us?"
	//
	// Get() uses the first header, but make sure to use the last header.
	xff := r.Header.Values("X-Forwarded-For")
	if len(xff) > 0 && xff[len(xff)-1] != "" {
		xffSplit := strings.Split(xff[len(xff)-1], ",")
		for i := len(xffSplit) - 1; i >= 0; i-- {
			if !znet.PrivateIPString(xffSplit[i]) {
				return strings.TrimSpace(xffSplit[i])
			}
		}
	}

	return znet.RemovePort(r.RemoteAddr)
}
