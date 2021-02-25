package zhttp

import (
	"net/http"
	"strings"

	"zgo.at/guru"
	"zgo.at/zstd/znet"
)

// HostRoute routes requests based on the Host header.
//
// The routers can be simple domain names ("example.com", "foo.example.com") or
// with a leading wildcard ("*.example.com", "*.foo.example.com").
//
// Exact matches are preferred over wildcards (e.g. "foo.example.com" trumps
// "*.example.com"). A single "*" will match any host and is used if nothing
// else matches.
func HostRoute(routers map[string]http.Handler) http.HandlerFunc {
	// Only one route which is a wildcard route, we can just pass everything
	// through to that.
	if r, ok := routers["*"]; ok && len(routers) == 1 {
		return r.ServeHTTP
	}

	wildcards := make(map[string]http.Handler)
	for k, v := range routers {
		if strings.HasPrefix(k, "*.") {
			wildcards[k[2:]] = v
		}
	}
	for k := range wildcards {
		delete(routers, "*."+k)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		host := znet.RemovePort(strings.ToLower(r.Host))
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
			ErrPage(w, r, guru.Errorf(404, "unknown domain: %q", r.Host))
			return
		}

		route.ServeHTTP(w, r)
	}
}

// RedirectHost redirects all requests to the destination host.
//
// Mainly intended for redirecting "example.com" to "www.example.com", or vice
// versa. The full URL is preserved, so "example.com/a?x is redirected to
// www.example.com/a?x
//
// Only GET requests are redirected.
func RedirectHost(dst string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Encode()
		if len(q) > 0 {
			q = "?" + q
		}
		w.Header().Set("Location", dst+r.URL.Path+q)
		w.WriteHeader(301)
	})
}
