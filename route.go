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
		route, ok := routers[strings.ToLower(r.Host)]
		if !ok {
			for k, v := range wildcards {
				if strings.HasSuffix(r.Host, k) {
					route = v
					ok = true
					break
				}
			}
		}

		if !ok {
			http.Error(w, fmt.Sprintf("unknown domain: %q", r.Host), http.StatusNotFound)
			return
		}

		route.ServeHTTP(w, r)
	}
}
