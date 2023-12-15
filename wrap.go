package zhttp

import (
	"context"
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
	"zgo.at/zlog"
	"zgo.at/ztpl"
)

// UserError modifies an error for user display.
//
//   - Removes any stack traces from zgo.at/errors or github.com/pkg/errors.
//   - Reformats some messages to be more user-friendly.
//   - "Hides" messages behind an error code for 5xx errors (you need to log those yourself).
func UserError(err error) (int, error) {
	if _, ok := err.(interface{ StackTrace() string }); ok {
		err = errors.Unwrap(err)
	}

	var (
		stErr interface {
			Code() int
			Error() string
		}
		dErr *ErrorDecode
		uErr *ErrorDecodeUnknown
		code = 500
	)
	switch {
	case errors.As(err, &stErr): // zgo.at/guru with an embedded status code.
		code = stErr.Code()
		err = stErr
	case errors.As(err, &dErr): // Invalid parameters.
		code = 400
	case errors.As(err, &uErr): // Invalid parameters.
		code = 400
	case errors.Is(err, sql.ErrNoRows):
		code = 404
	case errors.Is(err, context.DeadlineExceeded):
		code = http.StatusGatewayTimeout
	}

	switch {
	// Always use the same message for 404s; not just because it's easier but
	// also so that we're sure that an object which actually doesn't exist
	// appears the same as an object the user has no permissions to access,
	// which makes enumeration attacks harder.
	case code == 404:
		return code, errors.New("not found")
	case code == http.StatusGatewayTimeout:
		return code, errors.New("server timed out loading data")
	case code >= 500:
		return code, fmt.Errorf(
			`unexpected error code ‘%s’; this has been reported for investigation`,
			UserErrorCode(err))
	default:
		return code, err
	}
}

// UserErrorCode gets a hash of the error value.
func UserErrorCode(err error) string {
	if err == nil {
		return ""
	}

	h := fnv.New32()
	_, _ = h.Write([]byte(err.Error()))
	return strconv.FormatInt(int64(h.Sum32()), 36)
}

// ErrPage is the error page; this gets called whenever a "wrapped" handler
// returns an error.
//
// This is expected to write a status code and some content; for example a
// template with the error or some JSON data for an API request.
//
// nil errors should do nothing.
var ErrPage = DefaultErrPage

// DefaultErrPage is the default error page.
//
// Any unknown errors are displayed as an error code, with the real error being
// logged. This ensures people don't see entire stack traces or whatnot, which
// isn't too useful.
//
// The data written depends on the content type:
//
// Regular HTML GET requests try to render error.gohtml with ztpl if it's
// loaded, or writes a simple default HTML document instead. The Code and Error
// parameters are set for the HTML template.
//
// JSON requests write {"error" "the error message"}
//
// Forms add the error as a flash message and redirect back to the previous page
// (via the Referer header), or render the error.gohtml template if the header
// isn't set.
func DefaultErrPage(w http.ResponseWriter, r *http.Request, reported error) {
	if reported == nil {
		return
	}
	hasStatus := true
	if ww, ok := w.(statusWriter); !ok || ww.Status() == 0 {
		hasStatus = false
	}

	code, userErr := UserError(reported)
	if code >= 500 {
		zlog.Field("code", UserErrorCode(reported)).FieldsRequest(r).Error(reported)
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
		} else if jErr, ok := userErr.(interface{ ErrorJSON() ([]byte, error) }); ok {
			j, err = jErr.ErrorJSON()
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
			Path  string
		}{code, userErr, r.URL.Path})
		if err != nil {
			zlog.FieldsRequest(r).Error(err)
		}
	}
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Wrap a http.HandlerFunc and handle error returns.
func Wrap(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ErrPage(w, r, h(w, r))
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
