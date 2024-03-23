package mware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"zgo.at/zhttp"
	"zgo.at/zhttp/internal/isatty"
)

var enableColors = isatty.IsTerminal(os.Stdout.Fd())

type RequestLogOptions struct {
	Host    bool   // Print Host: as well
	TimeFmt string // Time format; default is just the time.
}

// RequestLog logs all requests to stdout.
//
// Any paths matching ignore will not be printed.
func RequestLog(opt *RequestLogOptions, ignore ...string) zhttp.Middleware {
	if opt == nil {
		opt = &RequestLogOptions{
			Host:    true,
			TimeFmt: "15:04:05 ",
		}
	}

	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			if len(ignore) > 0 {
				for _, i := range ignore {
					if r.URL.Path == i {
						return next(w, r)
					}
				}
			}

			start := time.Now()

			ww, ok := w.(zhttp.ResponseWriter)
			if !ok {
				ww = zhttp.NewResponseWriter(w, r.ProtoMajor)
			}
			nextErr := next(ww, r)

			// Get color-coded status code.
			status := "%d"
			if enableColors {
				switch {
				case ww.Status() == 0 || ww.Status() < 200 || ww.Status() >= 600:
					status = "INVALID STATUS CODE: '%d'"
				case ww.Status() >= 200 && ww.Status() < 400:
					status = "\x1b[48;5;154m\x1b[38;5;0m%d\x1b[0m"
				case ww.Status() >= 400 && ww.Status() <= 499:
					status = "\x1b[1m\x1b[48;5;221m\x1b[38;5;0m%d\x1b[0m"
				case ww.Status() >= 500 && ww.Status() <= 599:
					status = "\x1b[1m\x1b[48;5;9m\x1b[38;5;15m%d\x1b[0m"
				}
			}
			status = fmt.Sprintf(status, ww.Status())

			// Aligned method
			method := r.Method
			if len(method) < 5 {
				method = method + strings.Repeat(" ", 5-len(method))
			}

			url := r.URL.RequestURI()
			if opt.Host {
				url = r.Host + url
			}

			fmt.Printf("%s %s %s %3.0fms  %s\n", time.Now().Format(opt.TimeFmt),
				status, method, time.Since(start).Seconds()*1000, url)
			return nextErr
		})
	}
}

// WrapWriter replaces the http.ResponseWriter with our version of
// http.ResponseWriter for some additional functionality.
func WrapWriter() zhttp.Middleware {
	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return next(zhttp.NewResponseWriter(w, r.ProtoMajor), r)
		})
	}
}
