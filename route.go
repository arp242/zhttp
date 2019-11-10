package zhttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

// HostRoute redirects requests to chi.Routers based on the Host header.
//
// The routers can be simple domain names ("example.com", "foo.example.com") or
// with a leading wildcard ("*.example.com", "*.foo.example.com").
func HostRoute(routers map[string]chi.Router) http.HandlerFunc {
	wildcards := make(map[string]chi.Router)
	for k, v := range routers {
		if strings.HasPrefix(k, "*.") {
			wildcards[k[2:]] = v
		}
	}
	for k := range wildcards {
		delete(routers, "*."+k)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		host := RemovePort(strings.ToLower(r.Host))
		route, ok := routers[host]
		if !ok {
			for k, v := range wildcards {
				if strings.HasSuffix(host, k) {
					route = v
					ok = true
					break
				}
			}
		}
		if !ok {
			route, ok = routers["*"]
		}

		if !ok {
			ErrPage(w, r, 404, fmt.Errorf("unknown domain: %q", r.Host))
			return
		}

		route.ServeHTTP(w, r)
	}
}

// RedirectHost redirects all requests to the destination host.
//
// Mainly intended for redirecting "example.com" to "www.example.com", or vice
// verse. The full URL is preserved, so "example.com/a?x is redirected to
// www.example.com/a?x
//
// Only GET requests are redirected.
func RedirectHost(dst string) chi.Router {
	r := chi.NewRouter()
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Encode()
		if len(q) > 0 {
			q = "?" + q
		}
		w.Header().Set("Location", dst+r.URL.Path+q)
		w.WriteHeader(301)
	})
	return r
}
