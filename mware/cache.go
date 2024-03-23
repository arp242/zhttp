package mware

import (
	"net/http"

	"zgo.at/zhttp"
)

// NoCache sets the Cache-Control header to "no-cache".
//
// Browsers will always validate a cache (with e.g. If-Match or If-None-Match).
// It does NOT tell browsers to never store a cache (use NoStore for that).
func NoCache() zhttp.Middleware {
	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Cache-Control", "no-cache")
			return next(w, r)
		})
	}
}

// NoStore sets the Cache-Control header to "no-store, no-cache"
//
// Browsers will never store a local copy (the no-cache is there to be sure
// previously stored copies from before this header are revalidated).
func NoStore() zhttp.Middleware {
	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Cache-Control", "no-store,no-cache")
			return next(w, r)
		})
	}
}
