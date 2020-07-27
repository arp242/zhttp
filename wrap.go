package zhttp

import (
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"zgo.at/json"
	"zgo.at/zhttp/ztpl"
	"zgo.at/zlog"
)

type (
	coder interface {
		Code() int
		Error() string
	}
	errorJSON interface {
		ErrorJSON() ([]byte, error)
	}
	stackTracer interface {
		StackTrace() string
	}
)

// WrapWriter replaces the http.ResponseWriter with our version of
// http.ResponseWriter for some additional functionality.
func WrapWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(NewResponseWriter(w, r.ProtoMajor), r)
	})
}

var ErrPage = DefaultErrPage

func DefaultErrPage(w http.ResponseWriter, r *http.Request, code int, reported error) {
	hasStatus := true
	if ww, ok := w.(statusWriter); !ok || ww.Status() == 0 {
		hasStatus = false
	}

	userErr := reported
	if code >= 500 {
		zlog.Field("code", UserErrorCode(reported)).FieldsRequest(r).Error(reported)
		userErr = fmt.Errorf(
			`unexpected error code ‘%s’; this has been reported for investigation`,
			UserErrorCode(reported))
	}

	// Always use the same message for 404s; not just because it's easier but
	// also so that we're sure that an object which actually doesn't exist
	// appears the same as an object the user has no permissions to access,
	// which makes enumeration attacks harder.
	if code == 404 {
		userErr = errors.New("Not Found")
	}

	ct := strings.ToLower(r.Header.Get("Content-Type"))
	switch {
	case strings.HasPrefix(ct, "application/json"):
		if !hasStatus {
			w.WriteHeader(code)
		}

		var (
			j   []byte
			err error
		)

		if jErr, ok := userErr.(json.Marshaler); ok {
			j, err = jErr.MarshalJSON()
		} else if jErr, ok := userErr.(errorJSON); ok {
			j, err = jErr.ErrorJSON()
		} else if _, ok := userErr.(stackTracer); ok {
			j, err = json.Marshal(map[string]string{"error": errors.Unwrap(userErr).Error()})
		} else {
			j, err = json.Marshal(map[string]string{"error": userErr.Error()})
		}
		if err != nil {
			zlog.FieldsRequest(r).Error(err)
		}
		w.Write(j)

	case strings.HasPrefix(ct, "text/plain"):
		if !hasStatus {
			w.WriteHeader(code)
		}
		fmt.Fprintf(w, "Error %d: %s", code, userErr)

	case !hasStatus && r.Referer() != "" && ct == "application/x-www-form-urlencoded" || strings.HasPrefix(ct, "multipart/"):
		FlashError(w, userErr.Error())
		SeeOther(w, r.Referer())

	default:
		if !hasStatus {
			w.WriteHeader(code)
		}

		if !ztpl.HasTemplate("error.gohtml") {
			fmt.Fprintf(w, "<p>Error %d: %s</p>", code, userErr)
			return
		}

		err := ztpl.Execute(w, "error.gohtml", struct {
			Code  int
			Error error
		}{code, userErr})
		if err != nil {
			zlog.FieldsRequest(r).Error(err)
		}
	}
}

// UserErrorCode gets a hash based on the error value.
func UserErrorCode(err error) string {
	if err == nil {
		return ""
	}

	h := fnv.New32()
	_, _ = h.Write([]byte(err.Error()))
	return strconv.FormatInt(int64(h.Sum32()), 36)
}

// HandlerFunc function.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Wrap a http.HandlerFunc and handle error returns.
func Wrap(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		var (
			stErr coder
			dErr  *DecodeError
			code  = 500
		)
		switch {
		case errors.As(err, &stErr): // zgo.at/guru with an embedded status code.
			code = stErr.Code()
			err = stErr
		case errors.As(err, &dErr): // Invalid parameters.
			code = 400
		case errors.Is(err, sql.ErrNoRows):
			code = 404
		}
		ErrPage(w, r, code, err)
	}
}

func writeStatus(w http.ResponseWriter, code int, ct string) {
	if ww, ok := w.(statusWriter); !ok || ww.Status() == 0 {
		if ct != "" && w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(200)
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
func JSON(w http.ResponseWriter, i interface{}) error {
	var j []byte
	switch ii := i.(type) {
	case string:
		j = []byte(ii)
	case []byte:
		j = ii
	default:
		var err error
		j, err = json.MarshalIndent(i, "", "  ")
		if err != nil {
			return err
		}
	}

	writeStatus(w, 200, "application/json; charset=utf-8")
	w.Write(j)
	return nil
}

// Template renders a template and sends it to the client, setting the
// Content-Type to text/html (unless it's already set).
//
// This requires ztpl to be set up.
func Template(w http.ResponseWriter, name string, data interface{}) error {
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
