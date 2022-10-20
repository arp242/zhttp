package zhttp

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/monoculum/formam/v3"
	"zgo.at/guru"
	"zgo.at/json"
	"zgo.at/zlog"
)

const (
	ContentUnsupported uint8 = iota
	ContentQuery
	ContentForm
	ContentJSON
)

// LogUnknownFields tells Decode() to issue an error log on unknown fields.
var LogUnknownFields bool

type DecodeError struct {
	ct  uint8
	err error
}

func (e DecodeError) Unwrap() error { return e.err }
func (e DecodeError) Error() string {
	var s string
	switch e.ct {
	case ContentQuery:
		s += "invalid query parameters: "
	case ContentForm:
		s += "invalid form data: "
	case ContentJSON:
		s += "invalid JSON: "
	case ContentUnsupported:
		return "unsupported Content-Type"
	}
	return s + e.err.Error()
}

var formamOpts = &formam.DecoderOptions{
	TagName:     "json",
	TimeFormats: []string{"2006-01-02", time.RFC3339},
}

// Decode request parameters from a form, JSON body, or query parameters.
//
// Returns one of the Content* constants, which is useful if you want to
// alternate the responses.
func Decode(r *http.Request, dst interface{}) (uint8, error) {
	ct := r.Header.Get("Content-Type")
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}

	var (
		err error
		c   uint8
	)
	switch {
	case r.Method == http.MethodGet:
		c = ContentQuery
		err = formam.NewDecoder(formamOpts).Decode(r.URL.Query(), dst)
	case ct == "application/json":
		c = ContentJSON
		err = json.NewDecoder(r.Body).Decode(dst)
	case ct == "application/x-www-form-urlencoded":
		c = ContentForm
		err = r.ParseForm()
		if err == nil {
			err = formam.NewDecoder(formamOpts).Decode(r.Form, dst)
		}
	case ct == "multipart/form-data":
		c = ContentForm
		err = r.ParseMultipartForm(32 << 20) // 32MB, http.defaultMaxMemory
		if err == nil {
			err = formam.NewDecoder(formamOpts).Decode(r.Form, dst)
		}
	default:
		c = ContentUnsupported
		err = guru.Errorf(http.StatusUnsupportedMediaType,
			"unable to handle Content-Type %q", ct)
	}
	var fErr *formam.Error
	if errors.As(err, &fErr) && fErr.Code() == formam.ErrCodeUnknownField {
		if LogUnknownFields {
			zlog.FieldsRequest(r).Error(err)
		}
		err = nil
	}
	if err != nil && err != io.EOF {
		return 0, &DecodeError{c, err}
	}
	return c, nil
}
