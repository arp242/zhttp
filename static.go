package zhttp

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Static file server.
type Static struct {
	domain       string
	cacheControl map[string]int
	headers      map[string]map[string]string // Headers for specific URLs
	files        fs.FS
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
// The domain parameter is used for CORS.
//
// The Cache-Control header is set with the cache parameter, which is a path →
// cache mapping. The path is matched with filepath.Match() and the key "" is
// used if nothing matches. There is no guarantee about the order if multiple
// keys match. One of special Cache* constants can be used.
func NewStatic(domain string, files fs.FS, cache map[string]int) Static {
	for k := range cache {
		_, err := filepath.Match(k, "")
		if err != nil {
			panic(fmt.Sprintf("zhttp.NewStatic: invalid pattern in cache map: %s", err))
		}
	}

	return Static{domain: domain, cacheControl: cache, files: files, headers: make(map[string]map[string]string)}
}

var Static404 = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("file not found: %q", r.RequestURI), 404)
}

// Header sets headers for a path, overriding anything else.
func (s *Static) Header(path string, h map[string]string) {
	s.headers[path] = h
}

func (s Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "yeh nah", http.StatusForbidden)
		return
	}

	path := strings.TrimLeft(r.URL.Path, "/")
	d, err := fs.ReadFile(s.files, path)
	if err != nil {
		if os.IsNotExist(err) {
			Static404(w, r)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}

	ct := mime.TypeByExtension(filepath.Ext(path))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Access-Control-Allow-Origin", s.domain)
	if s.cacheControl != nil {
		cache := -100
		c, ok := s.cacheControl[r.URL.Path]
		if ok {
			cache = c
		}
		if cache == -100 {
			for k, v := range s.cacheControl {
				match, _ := filepath.Match(k, r.URL.Path)
				if match {
					cache = v
					break
				}
			}
		}
		if cache == -100 {
			cache = s.cacheControl["*"]
			if cache == 0 {
				cache = s.cacheControl[""]
			}
		}

		// TODO: use a more clever scheme where max-age and no-{cache,store} can be
		// set independently. For example, we can use the top 3 bits as a bitmask,
		// clear those if set, and use the rest as a max-age.
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
		if cc != "" {
			w.Header().Set("Cache-Control", cc)
		}
	}

	if h, ok := s.headers[r.URL.Path]; ok {
		for k, v := range h {
			w.Header().Set(k, v)
		}
	}

	w.WriteHeader(200)
	w.Write(d)
}
