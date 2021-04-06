// Package auth adds user authentication.

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"zgo.at/guru"
	"zgo.at/zhttp"
	"zgo.at/zhttp/ctxkey"
	"zgo.at/zstd/znet"
	"zgo.at/zstd/zstring"
)

var (
	cookieKey = "key"
	oneYear   = 24 * 365 * time.Hour
)

type loadFunc func(ctx context.Context, token string) (User, error)

type User interface {
	CSRFToken() string
}

// SetCookie sets the authentication cookie to val for the given domain.
func SetCookie(w http.ResponseWriter, val, domain string) {
	http.SetCookie(w, &http.Cookie{
		Domain:   znet.RemovePort(domain),
		Name:     cookieKey,
		Value:    val,
		Path:     "/",
		Expires:  time.Now().Add(oneYear),
		HttpOnly: true,
		Secure:   zhttp.CookieSecure,
		SameSite: zhttp.CookieSameSite,
	})
}

// ClearCookie sends an empty auth cookie with an expiry in the past for the
// given domain, clearing it in the client.
//
// Make sure the domain matches with what was sent before *exactly*, or the
// browser will set a second cookie.
func ClearCookie(w http.ResponseWriter, domain string) {
	http.SetCookie(w, &http.Cookie{
		Domain:  znet.RemovePort(domain),
		Name:    cookieKey,
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-24 * time.Hour),
	})
}

// Add user auth to an endpoint.
//
// The load callback is called with the value of the authentication cookie. The
// User is added to the context.
//
// POST, PATH, PUT, and DELETE requests will check the CSRF token from the
// "csrf" field with the value from User.CSRFToken(). The list of paths in
// noCSRF will be excluded for CSRF checks.
func Add(load loadFunc, noCSRF ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(cookieKey)
			if err != nil { // No cookie, no problem!
				// Ensure there's a concrete type (rather than nil) as that makes templating easier.
				u, _ := load(r.Context(), "")
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
				return
			}

			u, err := load(r.Context(), c.Value)
			if err != nil {
				// Clear cookies for both "foo.domain.com" and ".domain.com";
				// sometimes an invalid "stuck" cookie may be present,
				// preventing login. See https://github.com/zgoat/goatcounter/issues/387
				ClearCookie(w, r.Host)
				ClearCookie(w, strings.Join(strings.Split(r.Host, ".")[1:], "."))

				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
				return
			}

			// Check CSRF.
			if !zstring.Contains(noCSRF, r.URL.Path) {
				switch r.Method {
				case http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut:
					var err error
					if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
						err = r.ParseMultipartForm(32 << 20) // 32M, http.defaultMaxMemory
					} else {
						err = r.ParseForm()
					}
					if err != nil {
						w.WriteHeader(500)
						fmt.Fprintf(w, "zhttp.ParseForm: %s", err) // TODO: should probably use errpage?
						return
					}

					token := r.FormValue("csrf")
					r.Form.Del("csrf")
					if token == "" {
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprintln(w, "CSRF token is empty") // TODO: should probably use errpage?
						return
					} else {
						t := u.CSRFToken()
						if t != "" && token != t {
							w.WriteHeader(http.StatusForbidden)
							fmt.Fprintln(w, "Invalid CSRF token") // TODO: should probably use errpage?
							return
						}
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
		})
	}
}

type filterFunc func(http.ResponseWriter, *http.Request) error

// Filter access to a resource.
//
// If the returning error is a zgo.at/guru.coder and has a redirect code, then
// the error value is used as a redirection.
func Filter(f filterFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := f(w, r)
			if err == nil {
				next.ServeHTTP(w, r)
				return
			}

			code := http.StatusForbidden
			var cErr coder
			if errors.As(err, &cErr) {
				code = cErr.Code()
			}

			if code >= 300 && code <= 399 {
				w.Header().Set("Location", err.Error())
				w.WriteHeader(303)
				return
			}

			zhttp.ErrPage(w, r, guru.WithCode(code, err))
		})
	}
}

type coder interface {
	Code() int
	Error() string
}
