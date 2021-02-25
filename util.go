package zhttp

import (
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
