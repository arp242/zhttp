package rhttp // import "arp242.net/rhttp"

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// LogErrFunc gets called on errors.
var LogErrFunc func(error) = func(err error) {
	fmt.Fprintf(os.Stderr, "%s %s", time.Now().Format(time.RFC3339), err)
}

// HandlerFunc function.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Wrap a http.HandlerFunc
func Wrap(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)

		switch out := err.(type) {
		case nil:
			// Do nothing.

		// An actual error.
		default:
			LogErrFunc(out)
			w.WriteHeader(500)
			w.Write([]byte(out.Error()))
		}
	}
}

// SeeOther redirects to the given URL.
//
// "A 303 response to a GET request indicates that the origin server does not
// have a representation of the target resource that can be transferred by the
// server over HTTP. However, the Location field value refers to a resource that
// is descriptive of the target resource, such that making a retrieval request
// on that other resource might result in a representation that is useful to
// recipients without implying that it represents the original target resource."
func SeeOther(w http.ResponseWriter, url string) error {
	w.Header().Set("Location", url)
	w.WriteHeader(303)
	return nil
}
