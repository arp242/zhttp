package mhttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/monoculum/formam"
	"github.com/teamwork/guru"
)

// Decode a form or JSON paramters.
func Decode(r *http.Request, dst interface{}) error {
	ct := r.Header.Get("Content-Type")
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}

	var err error
	switch ct {
	case "application/json":
		err = json.NewDecoder(r.Body).Decode(dst)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		r.ParseForm()
		err = formam.NewDecoder(&formam.DecoderOptions{TagName: "json"}).Decode(r.Form, dst)
	default:
		err = guru.Errorf(http.StatusUnsupportedMediaType,
			"can only handle JSON and form requests, and not %q", ct)
	}
	return err
}
