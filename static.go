package zhttp

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Static file server.
type Static struct {
	dir          string
	domain       string
	cacheControl map[string]int
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
// The Cache-Control header is set with the cache parameter, which is a path →
// cache mapping. The path is matched with filepath.Match() and the key "" is
// used if nothing matches. There is no guarantee about the order if multiple
// keys match. One of special Cache* constants can be used.
//
// If packed is not nil then all files will be served from the map, and the
// filesystem is never accessed. This is useful for self-contained production builds.
//
// TODO: this should use http.FileSystem.
func NewStatic(dir, domain string, cache map[string]int, packed map[string][]byte) Static {
	for k := range cache {
		_, err := filepath.Match(k, "")
		if err != nil {
			panic(fmt.Sprintf("zhttp.NewStatic: invalid pattern in cache map: %s", err))
		}
	}

	return Static{dir: dir, domain: domain, cacheControl: cache, packed: packed}
}

var Static404 = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("packed file not found: %q", r.RequestURI), 404)
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
				Static404(w, r)
				return
			}
		}
		d = v
	} else {
		var err error
		d, err = ioutil.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				Static404(w, r)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
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
			cache = s.cacheControl[""]
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
	w.WriteHeader(200)
	w.Write(d)
}
