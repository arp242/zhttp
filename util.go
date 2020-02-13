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
