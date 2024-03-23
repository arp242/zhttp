package mware

import (
	"net/http"
	"strconv"
	"time"

	"zgo.at/zhttp"
	"zgo.at/zlog"
)

// Delay adds a delay before every request.
//
// The default delay is taken from the paramters (which may be 0), and can also
// be overriden by setting a "debug-delay" cookie, which is in seconds.
//
// This is intended for debugging frontend timing issues.
func Delay(d time.Duration) zhttp.Middleware {
	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
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
			return next(w, r)
		})
	}
}
