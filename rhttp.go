// Package rhttp
package rhttp // import "arp242.net/postit/rhttp"

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"
)

// LogErrFunc gets called on errors.
var LogErrFunc func(error) = func(err error) {
	fmt.Fprintf(os.Stderr, "%s %s", time.Now().Format(time.RFC3339), err)
}

// Default is called when no known error matches.
var Default = func(w http.ResponseWriter, err error) {
	LogErrFunc(err)
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

var ErrPage = func(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte("ERROR: "))
	w.Write([]byte(err.Error()))
}

// HandlerFunc function.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

type coder interface {
	Code() int
	Error() string
}

// Wrap a http.HandlerFunc
func Wrap(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		// A github.com/teamwork/guru error with an embeded status code.
		if stErr, ok := err.(coder); ok {
			ErrPage(w, r, stErr.Code(), stErr)
			return
		}

		switch err {
		case sql.ErrNoRows:
			ErrPage(w, r, 404, err)
			return
		}

		// switch out := err.(type) {
		// }

		Default(w, err)
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
