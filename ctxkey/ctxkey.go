// Package ctxkey stores context keys.
package ctxkey // import "zgo.at/zhttp/ctxkey"

// Context keys.
var (
	User = &struct{ n string }{"u"}
	Site = &struct{ n string }{"s"}
	DB   = &struct{ n string }{"d"}
)
