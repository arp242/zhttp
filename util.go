package zhttp

import (
	"net/http"
	"strings"
)

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

// IsSecure reports if this looks like a secure SSL/TLS connection.
//
// For proxied connections it depends on X-Forwarded-Proto being set. Most
// proxies set this by default.
func IsSecure(r *http.Request) bool {
	return !(r.TLS == nil || r.Header.Get("X-Forwarded-Proto") == "https")
}
