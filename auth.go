package zhttp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"zgo.at/zhttp/ctxkey"
)

var (
	cookieKey = "key"
	oneYear   = 24 * 365 * time.Hour
)

// Flags to add to all cookies (login and flash).
var (
	CookieSecure   = false
	CookieSameSite = http.SameSiteLaxMode
)

type loadFunc func(ctx context.Context, email string) (User, error)

type User interface {
	GetToken() string
}

// SetAuthCookie sets the authentication cookie to val for the given domain.
func SetAuthCookie(w http.ResponseWriter, val, domain string) {
	http.SetCookie(w, &http.Cookie{
		Domain:   RemovePort(domain),
		Name:     cookieKey,
		Value:    val,
		Path:     "/",
		Expires:  time.Now().Add(oneYear),
		HttpOnly: true,
		Secure:   CookieSecure,
		SameSite: CookieSameSite,
	})
}

// ClearAuthCookie sends an empty auth cookie with an expiry in the past for the
// given domain, clearing it in the client.
//
// Make sure the domain matches with what was sent before *exactly*, or the
// browser will set a second cookie.
func ClearAuthCookie(w http.ResponseWriter, domain string) {
	http.SetCookie(w, &http.Cookie{
		Domain:  RemovePort(domain),
		Name:    cookieKey,
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-24 * time.Hour),
	})
}

// Auth adds user auth to an endpoint.
func Auth(load loadFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(cookieKey)
			if err != nil { // No cooke, no problem!
				// Ensure there's a concrete type (rather than nil) as that makes templating easier.
				u, _ := load(r.Context(), "")
				//fmt.Fprintf(os.Stderr, "zhttp.Auth: no cookie: %#v\n", u) // TODO: log
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
				return
			}

			u, err := load(r.Context(), c.Value)
			if errors.Is(err, sql.ErrNoRows) { // Invalid token or whatever.
				//fmt.Fprintf(os.Stderr, "zhttp.Auth: no rows\n") // TODO: log
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
				return
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "zhttp.Auth: %s\n", err) // TODO: log
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
				return
			}

			// Check CSRF.
			switch r.Method {
			case http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut:
				var err error
				if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					err = r.ParseMultipartForm(32 << 20) // 32MB, http.defaultMaxMemory
				} else {
					err = r.ParseForm()
				}
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprintf(w, "zhttp.ParseForm: %s", err) // TODO: err
					return
				}

				token := r.FormValue("csrf")
				r.Form.Del("csrf")
				if token == "" {
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprintln(w, "CSRF token is empty")
					return
				} else {
					t := u.GetToken()
					if t != "" && token != t {
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprintln(w, "Invalid CSRF token")
						return
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

			ErrPage(w, r, code, err)
		})
	}
}
