package zhttp

import (
	"net/http"
	"strings"

	"zgo.at/guru"
	"zgo.at/zstd/znet"
)

// HostRoute routes requests based on the Host header.
//
// The routers can be simple domain names ("example.com", "foo.example.com"),
// with a leading "*." wildcard ("*.example.com", "*.foo.example.com"), or a
// trailing ".*" wildcard ("api.*", "foo.*").
//
// Exact matches are preferred over wildcards; e.g. "foo.example.com" trumps
// "*.example.com" or "foo.*". Leading wildcards take precedence over trailing
// wildcards.
//
// A single "*" matches any host and is used if nothing else matches.
func HostRoute(routers map[string]http.Handler) http.HandlerFunc {
	// There is only one route, which is a wildcard route. We can just pass
	// everything through to that.
	if r, ok := routers["*"]; ok && len(routers) == 1 {
		return r.ServeHTTP
	}

	var (
		postWild = make(map[string]http.Handler)
		preWild  = make(map[string]http.Handler)
	)
	for k, v := range routers {
		if strings.HasPrefix(k, "*.") {
			postWild[k[1:]] = v
		} else if strings.HasSuffix(k, ".*") {
			preWild[k[:len(k)-1]] = v
		}
	}
	for k := range postWild {
		delete(routers, "*."+k)
	}
	for k := range preWild {
		delete(routers, k+".*")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		host := znet.RemovePort(strings.ToLower(r.Host))
		route, ok := routers[host]
		if !ok {
			for k, v := range postWild {
				if strings.HasSuffix(host, k) {
					route = v
					ok = true
					break
				}
			}
		}
		if !ok {
			for k, v := range preWild {
				if strings.HasPrefix(host, k) {
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
// dst can start with http:// or https:// to force a specific protocol. If it's
// just a domain name or starts with "//" it will use the protocol client used
// (http:// or https://).
//
// A trailing ".*" will be replaced with the host; for example "www.*" will
// redirect "example.com" to "www.example.com" and "foo.example.com" to
// "www.foo.example.com".
//
// Only GET requests are redirected.
func RedirectHost(dst string) http.Handler {
	var copyProto, copyHost bool
	if strings.HasPrefix(dst, "//") {
		dst, copyProto = dst[2:], true
	}
	if !strings.HasPrefix(dst, "http://") && !strings.HasPrefix(dst, "https://") {
		copyProto = true
	}
	if strings.HasSuffix(dst, ".*") {
		dst, copyHost = dst[:len(dst)-1], true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst := dst
		if copyProto {
			if r.TLS != nil {
				dst = "https://" + dst
			} else {
				dst = "http://" + dst
			}
		}
		if copyHost {
			dst = dst + r.Host
		}

		path := r.URL.Path
		if path == "/" {
			path = ""
		}
		q := r.URL.Query().Encode()
		if len(q) > 0 {
			q = "?" + q
		}
		path += q

		w.Header().Set("Location", dst+path)
		w.WriteHeader(301)
	})
}
