package mware

import (
	"fmt"
	"net/http"

	"zgo.at/zhttp"
	"zgo.at/zstd/zdebug"
)

// Unpanic recovers from panics in handlers and calls ErrPage().
func Unpanic(filterStack ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", rec)
				}
				err = stackError{err, string(zdebug.Stack(append(filterStack,
					"net/http", "zgo.at/zhttp", "github.com/go-chi/chi")...))}
				zhttp.ErrPage(w, r, err)
			}()

			next.ServeHTTP(w, r)
		})
	}
}

type stackError struct {
	err   error
	stack string
}

func (s stackError) Unwrap() error      { return s.err }
func (s stackError) StackTrace() string { return s.stack }
func (s stackError) Error() string      { return s.stack + "\n" + s.err.Error() }
