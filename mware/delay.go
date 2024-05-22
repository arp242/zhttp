package mware

import (
	"net/http"
	"strconv"
	"time"

	"zgo.at/zlog"
)

// Delay adds a delay before every request.
//
// The default delay is taken from the paramters (which may be 0), and can also
// be overriden by setting a "debug-delay" cookie, which is in seconds.
//
// This is intended for debugging frontend timing issues.
func Delay(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			delay := d
			c, _ := r.Cookie("debug-delay")
			if c != nil {
				n, _ := strconv.ParseInt(c.Value, 10, 32)
				delay = time.Duration(n) * time.Second
			}

			if delay > 0 {
				zlog.Module("debug-delay").Printf("%s delay", delay)
				time.Sleep(delay)
			}
			next.ServeHTTP(w, r)
		})
	}
}
