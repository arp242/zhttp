package zhttp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
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
