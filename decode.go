package zhttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/monoculum/formam"
	"github.com/teamwork/guru"
)

const (
	ContentUnsupported uint8 = iota
	ContentForm
	ContentJSON
)

// Decode a form or JSON parameters.
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
	switch ct {
	case "application/json":
		c = ContentJSON
		err = json.NewDecoder(r.Body).Decode(dst)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		c = ContentForm
		r.ParseForm()
		err = formam.NewDecoder(&formam.DecoderOptions{TagName: "json"}).Decode(r.Form, dst)
	default:
		c = ContentUnsupported
		err = guru.Errorf(http.StatusUnsupportedMediaType,
			"can only handle JSON and form requests, and not %q", ct)
	}
	return c, err
}
