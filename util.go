package zhttp

import (
	"net"
	"strings"
)

// RemovePort removes the "port" part of an hostname.
//
// This only works for "host:port", and not URLs. See net.SplitHostPort.
func RemovePort(host string) string {
	shost, _, err := net.SplitHostPort(host)
	if err != nil { // Probably doesn't have a port
		return host
	}
	return shost
}

var tr = strings.NewReplacer(
	".", "",
	"/", "",
	`\`, "",
	"\x00", "",
)

// SafePath converts any string to a safe pathname, preventing directory
// traversal attacks and the like.
func SafePath(s string) string {
	return tr.Replace(s)
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
// This will return true for anything that is not an IP address, such as
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
