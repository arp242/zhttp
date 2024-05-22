package mware

import "net/http"

// TODO: put in zhttp
func With(handler http.Handler, wares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range wares {
		handler = m(handler)
	}
	return handler
}
