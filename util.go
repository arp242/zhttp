package zhttp

import "net"

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
