package zhttp

import (
	"net"
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
	cfip := r.Header.Get("Cf-Connecting-Ip")
	if cfip != "" && !PrivateIP(cfip) {
		return cfip
	}

	xrip := r.Header.Get("X-Real-Ip")
	if xrip != "" && !PrivateIP(xrip) {
		return xrip
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		xffSplit := strings.Split(xff, ",")
		for i := len(xffSplit) - 1; i >= 0; i-- {
			if !PrivateIP(xffSplit[i]) {
				return strings.TrimSpace(xffSplit[i])
			}
		}
	}

	return RemovePort(r.RemoteAddr)
}

var privateCIDR = func() []*net.IPNet {
	blocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs := make([]*net.IPNet, 0, len(blocks))
	for _, b := range blocks {
		_, cidr, _ := net.ParseCIDR(b)
		cidrs = append(cidrs, cidr)
	}
	return cidrs
}()

// PrivateIP reports if this is a private non-public IP address.
//
// This will also return true for anything that is not an IP address, such as
// "example.com" or "localhost".
func PrivateIP(ip string) bool {
	addr := net.ParseIP(RemovePort(strings.TrimSpace(ip)))
	if addr == nil { // Not an IP address?
		return true
	}

	for _, c := range privateCIDR {
		if c.Contains(addr) {
			return true
		}
	}
	return false
}
