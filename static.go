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
	dir    string
	domain string
	packed map[string][]byte
}

// NewStatic returns a new static fileserver.
//
// It will serve all files in dir. Symlinks are followed (even outside of dir!),
// but paths containing ".." are not allowed.
//
// The domain parameter is used for CORS.
//
// If packed is not nil then all files will be served from the map, and the
// filesystem is never accessed. This is useful for self-contained production builds.
func NewStatic(dir, domain string, packed map[string][]byte) Static {
	return Static{dir: dir, domain: domain, packed: packed}
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
			http.Error(w, fmt.Sprintf("packed file not found: %q", path), 404)
			return
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
	w.WriteHeader(200)
	w.Write(d)
}
