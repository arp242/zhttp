package zhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/monoculum/formam"
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
		err = formam.NewDecoder(&formam.DecoderOptions{TagName: "json"}).Decode(r.URL.Query(), dst)
	case ct == "application/json":
		c = ContentJSON
		err = json.NewDecoder(r.Body).Decode(dst)
	case ct == "application/x-www-form-urlencoded":
		c = ContentForm
		err = r.ParseForm()
		if err == nil {
			err = formam.NewDecoder(&formam.DecoderOptions{TagName: "json"}).Decode(r.Form, dst)
		}
	case ct == "multipart/form-data":
		c = ContentForm
		err = r.ParseMultipartForm(32 << 20) // 32MB, http.defaultMaxMemory
		if err == nil {
			err = formam.NewDecoder(&formam.DecoderOptions{TagName: "json"}).Decode(r.Form, dst)
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
	if err != nil {
		return 0, fmt.Errorf("zhttp.Decode: %w", err)
	}
	return c, nil
}
