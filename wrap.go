package zhttp

import (
	"io"
	"net/http"
	"os"

	"zgo.at/json"
	"zgo.at/ztpl"
)

type (
	// Handler with error return.
	Handler interface {
		ServeHTTP(http.ResponseWriter, *http.Request) error
	}

	// HandlerFunc with error return.
	HandlerFunc func(http.ResponseWriter, *http.Request) error

	// Middleware with error return.
	Middleware func(HandlerFunc) HandlerFunc

	// ServeMux works like [http.ServeMux], but uses error returns in the handler
	// functions.
	ServeMux struct{ *http.ServeMux }
)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// NewServeMux allocates and returns a new [ServeMux].
func NewServeMux() *ServeMux {
	return &ServeMux{http.NewServeMux()}
}

func (m *ServeMux) Handle(pattern string, handler Handler, middleware ...Middleware) {
	m.ServeMux.Handle(pattern, Wrap(handler, middleware...))
}

func (m *ServeMux) HandleFunc(pattern string, handler HandlerFunc, middleware ...Middleware) {
	m.ServeMux.HandleFunc(pattern, WrapFunc(handler, middleware...))
}

func (m *ServeMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	return m.ServeMux.Handler(r)
}

// Wrap a [http.Handler] and handle error returns.
func Wrap(h Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h.ServeHTTP)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ErrPage(w, r, h.ServeHTTP(w, r))
	})
}

// Wrap a http.HandlerFunc and handle error returns.
func WrapFunc(h HandlerFunc, middleware ...Middleware) http.HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ErrPage(w, r, h(w, r))
	}
}

// File outputs a file on the disk.
//
// This does NO PATH NORMALISATION! People can enter "../../../../etc/passwd".
// Make sure you sanitize your paths if they're from untrusted input.
func File(w http.ResponseWriter, path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	return Stream(w, fp)
}

// Stream any data to the client as-is.
func Stream(w http.ResponseWriter, fp io.Reader) error {
	writeStatus(w, 200, "")
	_, err := io.Copy(w, fp)
	return err
}

// Bytes sends the bytes as-is to the client.
func Bytes(w http.ResponseWriter, b []byte) error {
	writeStatus(w, 200, "")
	w.Write(b)
	return nil
}

// String sends the string as-is to the client.
func String(w http.ResponseWriter, s string) error {
	writeStatus(w, 200, "")
	w.Write([]byte(s))
	return nil
}

// Text sends the string to the client, setting the Content-Type to text/plain
// (unless it's already set).
func Text(w http.ResponseWriter, s string) error {
	writeStatus(w, 200, "text/plain; charset=utf-8")
	return String(w, s)
}

// JSON writes i as JSON, setting the Content-Type to application/json (unless
// it's already set).
//
// If i is a string or []byte it's assumed this is already JSON-encoded and
// sent it as-is rather than sending a JSON-fied string.
func JSON(w http.ResponseWriter, i any) error {
	var j []byte
	switch ii := i.(type) {
	case string:
		j = []byte(ii)
	case []byte:
		j = ii
	default:
		enc := json.NewEncoder(w)
		enc.NullArray(false)
		enc.SetIndent("", "  ")

		writeStatus(w, 200, "application/json; charset=utf-8")
		err := enc.Encode(i)
		if err != nil {
			return err
		}

		return nil
	}

	writeStatus(w, 200, "application/json; charset=utf-8")
	w.Write(j)
	return nil
}

// Template renders a template and sends it to the client, setting the
// Content-Type to text/html (unless it's already set).
//
// This requires ztpl to be set up.
func Template(w http.ResponseWriter, name string, data any) error {
	writeStatus(w, 200, "text/html; charset=utf-8")
	return ztpl.Execute(w, name, data)
}

// MovedPermanently redirects to the given URL with a 301.
func MovedPermanently(w http.ResponseWriter, url string) error {
	w.Header().Set("Location", url)
	w.WriteHeader(301)
	return nil
}

// SeeOther redirects to the given URL with a 303.
//
// "A 303 response to a GET request indicates that the origin server does not
// have a representation of the target resource that can be transferred by the
// server over HTTP. However, the Location field value refers to a resource that
// is descriptive of the target resource, such that making a retrieval request
// on that other resource might result in a representation that is useful to
// recipients without implying that it represents the original target resource."
func SeeOther(w http.ResponseWriter, url string) error {
	w.Header().Set("Location", url)
	w.WriteHeader(303)
	return nil
}

func writeStatus(w http.ResponseWriter, code int, ct string) {
	if ww, ok := w.(statusWriter); !ok || ww.Status() == 0 {
		if ct != "" && w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(200)
	}
}
