package zhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/monoculum/formam/v3"
	"zgo.at/guru"
	"zgo.at/json"
	"zgo.at/zlog"
)

type (
	ContentType uint8
)

func (c ContentType) String() string {
	switch c {
	default:
		return "Content-Type unsupported"
	case ContentQuery:
		return "Content-Type URL query"
	case ContentForm:
		return "Content-Type form"
	case ContentJSON:
		return "Content-Type JSON"
	}
}

const (
	ContentUnsupported ContentType = iota
	ContentQuery
	ContentForm
	ContentJSON
)

type (
	ErrorDecode struct {
		ct  ContentType
		err error
	}
	ErrorDecodeUnknown struct {
		Field string
	}
)

func (e ErrorDecodeUnknown) Error() string {
	return fmt.Sprintf("unknown parameter: %q", e.Field)
}

func (e ErrorDecode) Unwrap() error { return e.err }
func (e ErrorDecode) Error() string {
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

type Decoder struct {
	logUnknown bool
	retUnknown bool
}

// NewDecoder creates a new decoder.
//
// The default is to ignore unknown fields; if log is true an error log will be
// issued; if retErr is true they will be returned as a [ErrorDecodeUnknown].
func NewDecoder(log, retErr bool) Decoder {
	return Decoder{logUnknown: log, retUnknown: retErr}
}

// Decode request parameters from a form, JSON body, or query parameters.
//
// Returns one of the Content* constants, which is useful if you want to
// alternate the responses.
func (dec Decoder) Decode(r *http.Request, dst any) (ContentType, error) {
	ct := r.Header.Get("Content-Type")
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}

	var (
		err error
		c   ContentType
	)
	switch {
	case r.Method == http.MethodGet:
		c = ContentQuery
		err = formam.NewDecoder(formamOpts).Decode(r.URL.Query(), dst)
	case ct == "application/json":
		// TODO: UnknownFieldsLog isn't supported.
		d := json.NewDecoder(r.Body)
		if dec.retUnknown {
			d.DisallowUnknownFields()
		}

		c = ContentJSON
		err = d.Decode(dst)
		if err != nil && strings.HasPrefix(err.Error(), `json: unknown field "`) {
			_, field, _ := strings.Cut(err.Error(), `"`)
			return c, &ErrorDecodeUnknown{strings.Trim(field, `"`)}
		}
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
		if dec.logUnknown {
			zlog.FieldsRequest(r).Error(err)
		}
		if dec.retUnknown {
			return c, &ErrorDecodeUnknown{Field: fErr.Path()}
		}
		err = nil
	}
	if err != nil && err != io.EOF {
		return c, &ErrorDecode{c, err}
	}
	return c, nil
}

var DefaultDecoder = NewDecoder(false, false)

// Decode the request with [DefefaultDecoder].
func Decode(r *http.Request, dst any) (ContentType, error) {
	return DefaultDecoder.Decode(r, dst)
}
