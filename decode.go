package zhttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/monoculum/formam"
	"github.com/teamwork/guru"
	"zgo.at/zlog"
)

const (
	ContentUnsupported uint8 = iota
	ContentQuery
	ContentForm
	ContentJSON
)

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
	case ct == "application/x-www-form-urlencoded" || ct == "multipart/form-data":
		c = ContentForm
		err = r.ParseForm()
		if err == nil {
			err = formam.NewDecoder(&formam.DecoderOptions{TagName: "json"}).Decode(r.Form, dst)
		}
	default:
		c = ContentUnsupported
		err = guru.Errorf(http.StatusUnsupportedMediaType,
			"unable to handle Content-Type %q", ct)
	}
	if fErr, ok := err.(*formam.Error); ok && fErr.Code() == formam.ErrCodeUnknownField {
		zlog.Request(r).Error(err)
		err = nil
	}
	return c, err
}
