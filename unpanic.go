package zhttp

import (
	"fmt"
	"net/http"
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

				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("panic: %+v", rec)
				}

				ErrPage(w, r, 500, err)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
