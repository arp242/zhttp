package mware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"zgo.at/zhttp"
)

// Unpanic recovers from panics in handlers and calls ErrPage().
func Unpanic() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("panic at %s %s%s: %+v\n\nForm: %#v\nHeaders: %#v\n",
						r.Method, r.Host, r.RequestURI, rec, r.Form, r.Header)
				}

				err = fmt.Errorf("%w\n%s", err, debug.Stack())
				zhttp.ErrPage(w, r, 500, err)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
