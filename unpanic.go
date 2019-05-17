package mhttp

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

// TODO: https://github.com/Teamwork/middleware/blob/master/rescue/rescue.go
func Unpanic(prod bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				msg := fmt.Sprintf("Panic: %+v\n\n%s", rec, debug.Stack())
				fmt.Fprintf(os.Stderr, msg)

				if prod {
					msg = "Oops :-("
				}
				http.Error(w, msg, http.StatusInternalServerError)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
