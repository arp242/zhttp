// Package ctxkey stores context keys.
package ctxkey

// Context keys.
var (
	User = &struct{ n string }{"u"}
	Site = &struct{ n string }{"s"}
)
