package zhttp

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// Static file server.
type Static struct {
	dir          string
	domain       string
	cacheControl string
	packed       map[string][]byte
}

// Constants for the NewStatic() cache parameter.
const (
	// Don't set any header.
	CacheNoHeader = 0

	// Set to "no-cache" to tell browsers to always validate a cache (with e.g.
	// If-Match or If-None-Match). It does NOT tell browsers to never store a
	// cache; use Cache NoStore for that.
	CacheNoCache = -1

	// Set to "no-store, no-cache" to tell browsers to never store a local copy
	// (the no-cache is there to be sure previously stored copies from before
	// this header are revalidated).
	CacheNoStore = -2
)

// NewStatic returns a new static fileserver.
//
// It will serve all files in dir. Symlinks are followed (even outside of dir!),
// but paths containing ".." are not allowed.
//
// The domain parameter is used for CORS.
//
// Cache is set to the Cache-Control: max-age parameter, or use one of the
// special Cache* constants.
//
// If packed is not nil then all files will be served from the map, and the
// filesystem is never accessed. This is useful for self-contained production builds.
func NewStatic(dir, domain string, cache int, packed map[string][]byte) Static {
	cc := ""
	switch cache {
	case CacheNoHeader:
		cc = ""
	case CacheNoCache:
		cc = "no-cache"
	case CacheNoStore:
		cc = "no-store,no-cache"
	default:
		cc = fmt.Sprintf("public, max-age=%d", cache)
	}

	return Static{dir: dir, domain: domain, cacheControl: cc, packed: packed}
}

func (s Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "yeh nah", http.StatusForbidden)
		return
	}

	path := filepath.Join(s.dir, "/", r.URL.Path)

	var d []byte
	if s.packed != nil {
		v, ok := s.packed[path]
		if !ok {
			path = path + "/index.html"
			v, ok = s.packed[path]
			if !ok {
				http.Error(w, fmt.Sprintf("packed file not found: %q", path), 404)
				return
			}
		}
		d = v
	} else {
		var err error
		d, err = ioutil.ReadFile(path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	w.Header().Set("Access-Control-Allow-Origin", s.domain)
	if s.cacheControl != "" {
		w.Header().Set("Cache-Control", s.cacheControl)
	}
	w.WriteHeader(200)
	w.Write(d)
}
