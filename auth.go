package zhttp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/teamwork/utils/netutil"

	"zgo.at/zhttp/ctxkey"
)

var (
	cookieKey = "key"
	oneYear   = 24 * 365 * time.Hour
)

type loadFunc func(ctx context.Context, email string) (User, error)

type User interface {
	GetToken() string
}

func SetCookie(w http.ResponseWriter, val, domain string) {
	http.SetCookie(w, &http.Cookie{
		Domain:  netutil.RemovePort(domain),
		Name:    cookieKey,
		Value:   val,
		Path:    "/",
		Expires: time.Now().Add(oneYear),
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieKey,
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-24 * time.Hour),
	})
}

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
			if errors.Cause(err) == sql.ErrNoRows { // Invalid token or whatever.
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
				err := r.ParseForm()
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprintf(w, "ParseForm: %s", err) // TODO: err
					return
				}

				token := r.FormValue("csrf")
				r.Form.Del("csrf")
				if token == "" {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, "token is empty") // TODO: err
					return
				}

				t := u.GetToken()
				if t != "" && token != t {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Printf("token %q doesn't match strored %q\n", token, t)
					fmt.Fprintf(w, "token doesn't match")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxkey.User, u)))
		})
	}
}

type filterFunc func(http.ResponseWriter, *http.Request) error

// Filter access to a resource.
//
// If the returning error is a github.com/teamwork/guru.coder and has a redirect
// code, then the error value is used as a redirection.
func Filter(f filterFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := f(w, r)
			if err == nil {
				next.ServeHTTP(w, r)
				return
			}

			code := http.StatusForbidden
			if cErr, ok := err.(coder); ok {
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
