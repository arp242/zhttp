// Package mhttp
package mhttp // import "arp242.net/stentor/mhttp"

import (
	"database/sql"
	"fmt"
	"net/http"
)

var ErrPage = func(w http.ResponseWriter, r *http.Request, code int, err error) {
	fmt.Println("ErrPage", code, err)
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

		ErrPage(w, r, 500, err)
	}
}

func String(w http.ResponseWriter, s string) error {
	w.Write([]byte(s))
	w.WriteHeader(303)
	return nil
}

func Template(w http.ResponseWriter, name string, data interface{}) error {
	return tpl.ExecuteTemplate(w, name, data)
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
